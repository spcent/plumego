package ipc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func ExampleServer() {
	// Create server with custom config
	config := &Config{
		ConnectTimeout: 5 * time.Second,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		BufferSize:     8192,
	}

	server, err := NewServerWithConfig("test_socket", config)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	fmt.Printf("Server listening on: %s\n", server.Addr())

	// Accept connections
	go func() {
		for {
			client, acceptErr := server.Accept()
			if acceptErr != nil {
				log.Printf("Accept error: %v", acceptErr)
				return
			}

			// Handle client in goroutine
			go handleClient(client)
		}
	}()

	// Client connection example
	client, err := DialWithConfig(server.Addr(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Send data with timeout
	data := []byte("Hello, IPC!")
	n, err := client.WriteWithTimeout(data, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent %d bytes\n", n)

	// Read response
	buf := make([]byte, 1024)
	n, err = client.ReadWithTimeout(buf, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received: %s\n", string(buf[:n]))
}

func handleClient(client Client) {
	defer client.Close()

	buf := make([]byte, 1024)
	for {
		n, err := client.Read(buf)
		if err != nil {
			return
		}

		// Echo back the data
		client.Write(buf[:n])
	}
}

// TestConfig tests the configuration system
func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultConfig()
		if config == nil {
			t.Fatal("DefaultConfig returned nil")
		}
		if config.ConnectTimeout != 10*time.Second {
			t.Errorf("Expected ConnectTimeout 10s, got %v", config.ConnectTimeout)
		}
		if config.BufferSize != 4096 {
			t.Errorf("Expected BufferSize 4096, got %d", config.BufferSize)
		}
	})
}

// TestServerBasic tests basic server functionality
func TestServerBasic(t *testing.T) {
	t.Run("CreateAndClose", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		addr := server.Addr()
		if addr == "" {
			t.Error("Server address is empty")
		}

		if err := server.Close(); err != nil {
			t.Errorf("Failed to close server: %v", err)
		}
	})

	t.Run("WithCustomConfig", func(t *testing.T) {
		config := &Config{
			ConnectTimeout: 5 * time.Second,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			BufferSize:     8192,
		}

		server, err := NewServerWithConfig(getTestAddr(), config)
		if err != nil {
			t.Fatalf("Failed to create server with config: %v", err)
		}
		defer server.Close()

		if server.Addr() == "" {
			t.Error("Server address is empty")
		}
	})

	t.Run("MultipleClose", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Close multiple times should not error
		if err := server.Close(); err != nil {
			t.Errorf("First close failed: %v", err)
		}
		if err := server.Close(); err != nil {
			t.Errorf("Second close failed: %v", err)
		}
	})
}

// TestClientServer tests client-server communication
func TestClientServer(t *testing.T) {
	t.Run("BasicCommunication", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()
		defer client.Close()

		testData := []byte("Hello, IPC!")

		// Write from client
		n, err := client.Write(testData)
		if err != nil {
			t.Fatalf("Client write failed: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}

		// Accept and read from server
		serverClient, err := server.Accept()
		if err != nil {
			t.Fatalf("Server accept failed: %v", err)
		}
		defer serverClient.Close()

		buf := make([]byte, 1024)
		n, err = serverClient.Read(buf)
		if err != nil {
			t.Fatalf("Server read failed: %v", err)
		}

		received := buf[:n]
		if string(received) != string(testData) {
			t.Errorf("Expected %q, got %q", testData, received)
		}
	})

	t.Run("BidirectionalCommunication", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()
		defer client.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		// Server goroutine
		go func() {
			defer wg.Done()
			serverClient, err := server.Accept()
			if err != nil {
				t.Errorf("Server accept failed: %v", err)
				return
			}
			defer serverClient.Close()

			// Read from client
			buf := make([]byte, 1024)
			n, err := serverClient.Read(buf)
			if err != nil {
				t.Errorf("Server read failed: %v", err)
				return
			}

			// Echo back with prefix
			response := append([]byte("Echo: "), buf[:n]...)
			if _, err := serverClient.Write(response); err != nil {
				t.Errorf("Server write failed: %v", err)
			}
		}()

		// Client goroutine
		go func() {
			defer wg.Done()

			// Wait a bit for server to be ready
			time.Sleep(100 * time.Millisecond)

			testData := []byte("Hello from client")
			if _, err := client.Write(testData); err != nil {
				t.Errorf("Client write failed: %v", err)
				return
			}

			// Read response
			buf := make([]byte, 1024)
			n, err := client.Read(buf)
			if err != nil {
				t.Errorf("Client read failed: %v", err)
				return
			}

			expected := "Echo: Hello from client"
			if string(buf[:n]) != expected {
				t.Errorf("Expected %q, got %q", expected, string(buf[:n]))
			}
		}()

		wg.Wait()
	})

	t.Run("MultipleClients", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		numClients := 5
		var wg sync.WaitGroup
		wg.Add(numClients)
		var acceptWG sync.WaitGroup
		acceptWG.Add(numClients)

		// Start clients
		for i := 0; i < numClients; i++ {
			go func(clientID int) {
				defer wg.Done()

				client, err := Dial(server.Addr())
				if err != nil {
					t.Errorf("Client %d dial failed: %v", clientID, err)
					return
				}
				defer client.Close()

				data := fmt.Sprintf("Client %d data", clientID)
				if _, err := client.Write([]byte(data)); err != nil {
					t.Errorf("Client %d write failed: %v", clientID, err)
				}
			}(i)
		}

		// Accept clients on server
		for i := 0; i < numClients; i++ {
			go func() {
				defer acceptWG.Done()
				serverClient, err := server.Accept()
				if err != nil {
					// Ignore closed server during shutdown
					if errors.Is(err, net.ErrClosed) {
						return
					}
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					t.Errorf("Server accept failed: %v", err)
					return
				}
				defer serverClient.Close()

				buf := make([]byte, 1024)
				if _, err := serverClient.Read(buf); err != nil {
					if !errors.Is(err, io.EOF) {
						t.Errorf("Server read failed: %v", err)
					}
				}
			}()
		}

		wg.Wait()
		acceptWG.Wait()
	})
}

// TestWithTimeout tests timeout functionality
func TestWithTimeout(t *testing.T) {
	t.Run("WriteTimeout", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()
		defer client.Close()

		// This should succeed quickly
		data := []byte("test data")
		n, err := client.WriteWithTimeout(data, 1*time.Second)
		if err != nil {
			t.Errorf("WriteWithTimeout failed: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}
	})

	t.Run("ReadTimeout", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()
		defer client.Close()

		// Start a server that delays response
		go func() {
			serverClient, err := server.Accept()
			if err != nil {
				return
			}
			defer serverClient.Close()

			// Read but don't respond immediately
			buf := make([]byte, 1024)
			serverClient.Read(buf)

			// Delay response
			time.Sleep(200 * time.Millisecond)
			serverClient.Write([]byte("delayed response"))
		}()

		// Send data
		client.Write([]byte("test"))

		// Try to read with short timeout (should work since we delay only 200ms)
		buf := make([]byte, 1024)
		_, err := client.ReadWithTimeout(buf, 500*time.Millisecond)
		if err != nil {
			t.Errorf("ReadWithTimeout should have succeeded: %v", err)
		}
	})

	t.Run("ConnectTimeout", func(t *testing.T) {
		config := &Config{
			ConnectTimeout: 100 * time.Millisecond,
			ReadTimeout:    1 * time.Second,
			WriteTimeout:   1 * time.Second,
			BufferSize:     4096,
		}

		// Try to connect to non-existent address
		_, err := DialWithConfig("127.0.0.1:0", config)
		if err == nil {
			t.Error("Expected connect timeout error")
		}
	})
}

// TestWithContext tests context functionality
func TestWithContext(t *testing.T) {
	t.Run("AcceptWithContext", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// This should timeout since no client connects
		_, err = server.AcceptWithContext(ctx)
		if err == nil {
			t.Error("Expected context timeout error")
		}
	})

	t.Run("DialWithContext", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		client, err := DialWithContext(ctx, server.Addr(), nil)
		if err != nil {
			t.Fatalf("DialWithContext failed: %v", err)
		}
		defer client.Close()

		if client.RemoteAddr() == "" {
			t.Error("Client RemoteAddr is empty")
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		ctx, cancel := context.WithCancel(context.Background())

		// Cancel immediately
		cancel()

		_, err = DialWithContext(ctx, server.Addr(), nil)
		if err == nil {
			t.Error("Expected context cancellation error")
		}
	})
}

// TestErrorConditions tests various error conditions
func TestErrorConditions(t *testing.T) {
	t.Run("WriteToClosedClient", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()

		client.Close()

		_, err := client.Write([]byte("test"))
		if err == nil {
			t.Error("Expected error writing to closed client")
		}
	})

	t.Run("ReadFromClosedClient", func(t *testing.T) {
		server, client := setupTestPair(t)
		defer server.Close()

		client.Close()

		buf := make([]byte, 1024)
		_, err := client.Read(buf)
		if err == nil {
			t.Error("Expected error reading from closed client")
		}
	})

	t.Run("AcceptFromClosedServer", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		server.Close()

		_, err = server.Accept()
		if err == nil {
			t.Error("Expected error accepting from closed server")
		}
	})

	t.Run("DialNonExistentServer", func(t *testing.T) {
		_, err := Dial("127.0.0.1:1") // Port 1 should be unavailable
		if err == nil {
			t.Error("Expected error dialing non-existent server")
		}
	})
}

// TestConcurrency tests concurrent operations
func TestConcurrency(t *testing.T) {
	t.Run("ConcurrentReadsWrites", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		numRoutines := 5
		var clientWG sync.WaitGroup
		var serverWG sync.WaitGroup
		clientWG.Add(numRoutines)
		serverWG.Add(numRoutines)

		for i := 0; i < numRoutines; i++ {
			go func() {
				defer serverWG.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				serverClient, err := server.AcceptWithContext(ctx)
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return
					}
					t.Errorf("AcceptWithContext failed: %v", err)
					return
				}
				defer serverClient.Close()

				buf := make([]byte, 64)
				if _, err := serverClient.ReadWithTimeout(buf, time.Second); err != nil {
					if nerr, ok := err.(net.Error); !ok || !nerr.Timeout() {
						t.Errorf("ReadWithTimeout failed: %v", err)
					}
				}
			}()
		}

		for i := 0; i < numRoutines; i++ {
			go func(id int) {
				defer clientWG.Done()
				client, err := Dial(server.Addr())
				if err != nil {
					t.Errorf("Dial failed: %v", err)
					return
				}
				defer client.Close()

				data := fmt.Sprintf("client-%d", id)
				if _, err := client.WriteWithTimeout([]byte(data), time.Second); err != nil {
					t.Errorf("WriteWithTimeout failed: %v", err)
				}
			}(i)
		}

		clientWG.Wait()
		serverWG.Wait()
	})

	t.Run("ConcurrentAccepts", func(t *testing.T) {
		server, err := NewServer(getTestAddr())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}
		defer server.Close()

		numClients := 5
		var wg sync.WaitGroup
		wg.Add(numClients * 2) // clients + server accepts

		// Start accept goroutines
		for i := 0; i < numClients; i++ {
			go func() {
				defer wg.Done()
				serverClient, err := server.Accept()
				if err != nil {
					t.Errorf("Accept failed: %v", err)
					return
				}
				serverClient.Close()
			}()
		}

		// Start client connections
		for i := 0; i < numClients; i++ {
			go func() {
				defer wg.Done()
				client, err := Dial(server.Addr())
				if err != nil {
					t.Errorf("Dial failed: %v", err)
					return
				}
				client.Close()
			}()
		}

		wg.Wait()
	})
}

// TestLargeData tests handling of large data transfers
func TestLargeData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data test in short mode")
	}

	server, client := setupTestPair(t)
	defer server.Close()
	defer client.Close()

	serverClient, err := server.Accept()
	if err != nil {
		t.Fatalf("Server accept failed: %v", err)
	}
	defer serverClient.Close()

	// Test with 1MB of data
	dataSize := 1024 * 1024
	testData := make([]byte, dataSize)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var readData []byte
	var readErr error

	// Server reads all data
	go func() {
		defer wg.Done()
		readData = make([]byte, 0, dataSize)
		buf := make([]byte, 4096)

		for len(readData) < dataSize {
			n, err := serverClient.Read(buf)
			if err != nil {
				if err != io.EOF {
					readErr = err
				}
				break
			}
			readData = append(readData, buf[:n]...)
		}
	}()

	// Client sends all data
	go func() {
		defer wg.Done()

		sent := 0
		for sent < len(testData) {
			n, err := client.Write(testData[sent:])
			if err != nil {
				t.Errorf("Write failed: %v", err)
				break
			}
			sent += n
		}
	}()

	wg.Wait()

	if readErr != nil {
		t.Fatalf("Read error: %v", readErr)
	}

	if len(readData) != len(testData) {
		t.Errorf("Expected %d bytes, got %d", len(testData), len(readData))
	}

	// Verify data integrity
	for i, b := range testData {
		if i >= len(readData) || readData[i] != b {
			t.Errorf("Data mismatch at position %d: expected %d, got %d", i, b, readData[i])
			break
		}
	}
}

// Benchmark tests
func BenchmarkClientServer(b *testing.B) {
	server, client := setupTestPairB(b)
	defer server.Close()
	defer client.Close()

	serverClient, err := server.Accept()
	if err != nil {
		b.Fatalf("Server accept failed: %v", err)
	}
	defer serverClient.Close()

	data := []byte("benchmark data")
	buf := make([]byte, 1024)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.Write(data)
			serverClient.Read(buf)
		}
	})
}

func BenchmarkLargeWrite(b *testing.B) {
	server, client := setupTestPairB(b)
	defer server.Close()
	defer client.Close()

	serverClient, err := server.Accept()
	if err != nil {
		b.Fatalf("Server accept failed: %v", err)
	}
	defer serverClient.Close()

	data := make([]byte, 64*1024) // 64KB
	buf := make([]byte, 64*1024)

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		go serverClient.Read(buf)
		client.Write(data)
	}
}

// Helper functions

func getTestAddr() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("test_socket_%d_%d", os.Getpid(), time.Now().UnixNano())
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("test_socket_%d_%d", os.Getpid(), time.Now().UnixNano()))
}

func setupTestPair(t *testing.T) (Server, Client) {
	server, err := NewServer(getTestAddr())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	client, err := Dial(server.Addr())
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}

	return server, client
}

func setupTestPairB(b *testing.B) (Server, Client) {
	addr := getTestAddrB()
	server, err := NewServer(addr)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	client, err := Dial(server.Addr())
	if err != nil {
		server.Close()
		b.Fatalf("Failed to create client: %v", err)
	}

	return server, client
}

func getTestAddrB() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("bench_socket_%d_%d", os.Getpid(), time.Now().UnixNano())
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("bench_socket_%d_%d", os.Getpid(), time.Now().UnixNano()))
}
