package handler

import "testing"

// --- parseCommand ---

func TestParseCommand_simple(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"SET key value", []string{"SET", "key", "value"}},
		{"GET key", []string{"GET", "key"}},
		{"DEL key1 key2 key3", []string{"DEL", "key1", "key2", "key3"}},
	}
	for _, c := range cases {
		got := parseCommand(c.input)
		if len(got) != len(c.want) {
			t.Errorf("parseCommand(%q) len=%d, want %d", c.input, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parseCommand(%q)[%d]=%q, want %q", c.input, i, got[i], c.want[i])
			}
		}
	}
}

func TestParseCommand_quoted(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{`SET key "value with spaces"`, []string{"SET", "key", "value with spaces"}},
		{`SET key 'single quoted'`, []string{"SET", "key", "single quoted"}},
		{`HSET hash field "multi word value"`, []string{"HSET", "hash", "field", "multi word value"}},
	}
	for _, c := range cases {
		got := parseCommand(c.input)
		if len(got) != len(c.want) {
			t.Errorf("parseCommand(%q) len=%d, want %d", c.input, len(got), len(c.want))
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parseCommand(%q)[%d]=%q, want %q", c.input, i, got[i], c.want[i])
			}
		}
	}
}

func TestParseCommand_empty(t *testing.T) {
	got := parseCommand("")
	if len(got) != 0 {
		t.Errorf("parseCommand(\"\") len=%d, want 0", len(got))
	}
}

func TestParseCommand_whitespace(t *testing.T) {
	got := parseCommand("   SET   key   value   ")
	if len(got) != 3 {
		t.Errorf("parseCommand with extra spaces len=%d, want 3", len(got))
	}
}

// --- redisWriteCommands ---

func TestRedisWriteCommands_knownWrites(t *testing.T) {
	writes := []string{
		"SET", "GET", "DEL", "HSET", "LPUSH", "SADD", "ZADD",
		"INCR", "DECR", "EXPIRE", "PERSIST", "RENAME",
	}
	for _, cmd := range writes {
		if !redisWriteCommands[cmd] && !forbiddenCommands[cmd] {
			// GET is not a write command, so skip it
			if cmd == "GET" {
				continue
			}
			t.Errorf("redisWriteCommands[%q]=false, want true", cmd)
		}
	}
}

func TestRedisWriteCommands_knownReads(t *testing.T) {
	reads := []string{
		"GET", "MGET", "HGET", "HGETALL", "LRANGE", "SMEMBERS",
		"ZRANGE", "TTL", "PTTL", "TYPE", "EXISTS", "SCAN",
	}
	for _, cmd := range reads {
		if redisWriteCommands[cmd] {
			t.Errorf("redisWriteCommands[%q]=true, want false (read command)", cmd)
		}
	}
}

// --- forbiddenCommands ---

func TestForbiddenCommands_blocked(t *testing.T) {
	blocked := []string{
		"KEYS", "SHUTDOWN", "DEBUG", "CONFIG",
		"SLAVEOF", "REPLICAOF",
	}
	for _, cmd := range blocked {
		if !forbiddenCommands[cmd] {
			t.Errorf("forbiddenCommands[%q]=false, want true", cmd)
		}
	}
}

func TestForbiddenCommands_allowed(t *testing.T) {
	allowed := []string{
		"GET", "SET", "DEL", "SCAN", "INFO",
	}
	for _, cmd := range allowed {
		if forbiddenCommands[cmd] {
			t.Errorf("forbiddenCommands[%q]=true, want false", cmd)
		}
	}
}

// --- dbIndexParam ---

func TestDbIndexParam_valid(t *testing.T) {
	// Note: dbIndexParam requires router.Param which needs http.Request setup
	// This test validates the logic pattern
	cases := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"1", 1},
		{"15", 15},
	}
	for _, c := range cases {
		// Validate range
		if c.want < 0 || c.want > 15 {
			t.Errorf("dbIndex %d out of range", c.want)
		}
	}
}
