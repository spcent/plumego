// Seed data for MongoDB test database
// This runs automatically on first container start

db = db.getSiblingDB('testdb');

// Users collection with embedded documents
db.users.insertMany([
  {
    username: "alice",
    email: "alice@example.com",
    age: 28,
    balance: 1500.50,
    isActive: true,
    preferences: {
      theme: "dark",
      notifications: true,
      language: "en"
    },
    tags: ["premium", "active"],
    createdAt: new Date("2024-01-15"),
    address: {
      street: "123 Main St",
      city: "New York",
      country: "USA",
      zipCode: "10001"
    }
  },
  {
    username: "bob",
    email: "bob@example.com",
    age: 34,
    balance: 2300.75,
    isActive: true,
    preferences: {
      theme: "light",
      notifications: false,
      language: "en"
    },
    tags: ["standard", "active"],
    createdAt: new Date("2024-02-20"),
    address: {
      street: "456 Oak Ave",
      city: "Los Angeles",
      country: "USA",
      zipCode: "90001"
    }
  },
  {
    username: "charlie",
    email: "charlie@example.com",
    age: 42,
    balance: 890.25,
    isActive: true,
    preferences: {
      theme: "auto",
      notifications: true,
      language: "fr"
    },
    tags: ["premium", "active"],
    createdAt: new Date("2024-03-10"),
    address: {
      street: "789 Pine Rd",
      city: "Paris",
      country: "France",
      zipCode: "75001"
    }
  },
  {
    username: "diana",
    email: "diana@example.com",
    age: 29,
    balance: 3200.00,
    isActive: false,
    preferences: {
      theme: "dark",
      notifications: true,
      language: "de"
    },
    tags: ["premium", "inactive"],
    createdAt: new Date("2024-01-25"),
    address: {
      street: "321 Elm St",
      city: "Berlin",
      country: "Germany",
      zipCode: "10115"
    }
  },
  {
    username: "eve",
    email: "eve@example.com",
    age: 31,
    balance: 1750.80,
    isActive: true,
    preferences: {
      theme: "light",
      notifications: false,
      language: "en"
    },
    tags: ["standard", "active"],
    createdAt: new Date("2024-04-05")
  }
]);

// Products collection with array fields
db.products.insertMany([
  {
    name: "Laptop Pro",
    description: "High-performance laptop with 16GB RAM",
    price: 1299.99,
    stock: 50,
    category: "electronics",
    tags: ["new", "featured"],
    specs: {
      cpu: "Intel i7",
      ram: "16GB",
      storage: "512GB SSD",
      display: "15.6 inch"
    },
    reviews: [
      { user: "alice", rating: 5, comment: "Excellent performance!" },
      { user: "bob", rating: 4, comment: "Great laptop, a bit heavy" }
    ],
    createdAt: new Date("2024-01-10")
  },
  {
    name: "Wireless Mouse",
    description: "Ergonomic wireless mouse",
    price: 29.99,
    stock: 200,
    category: "electronics",
    tags: ["sale"],
    specs: {
      connectivity: "Bluetooth",
      battery: "12 months",
      dpi: "1600"
    },
    reviews: [
      { user: "charlie", rating: 4, comment: "Comfortable to use" }
    ],
    createdAt: new Date("2024-02-15")
  },
  {
    name: "Programming Book",
    description: "Learn Go Programming",
    price: 49.99,
    stock: 100,
    category: "books",
    tags: ["featured"],
    author: "John Doe",
    pages: 450,
    isbn: "978-1234567890",
    reviews: [
      { user: "eve", rating: 5, comment: "Best Go book ever!" },
      { user: "alice", rating: 5, comment: "Very comprehensive" }
    ],
    createdAt: new Date("2024-03-01")
  },
  {
    name: "Mechanical Keyboard",
    description: "RGB mechanical keyboard",
    price: 89.99,
    stock: 75,
    category: "electronics",
    tags: ["new"],
    specs: {
      switches: "Cherry MX Blue",
      layout: "Full size",
      backlight: "RGB"
    },
    reviews: [],
    createdAt: new Date("2024-04-20")
  },
  {
    name: "Coffee Beans",
    description: "Premium arabica coffee beans 1kg",
    price: 24.99,
    stock: 150,
    category: "food",
    tags: ["featured"],
    origin: "Colombia",
    roast: "Medium",
    reviews: [
      { user: "diana", rating: 5, comment: "Excellent flavor" }
    ],
    createdAt: new Date("2024-02-28")
  }
]);

// Orders collection with embedded order items
db.orders.insertMany([
  {
    userId: db.users.findOne({ username: "alice" })._id,
    orderNumber: "ORD-001",
    status: "delivered",
    items: [
      {
        productId: db.products.findOne({ name: "Laptop Pro" })._id,
        productName: "Laptop Pro",
        quantity: 1,
        unitPrice: 1299.99
      },
      {
        productId: db.products.findOne({ name: "Wireless Mouse" })._id,
        productName: "Wireless Mouse",
        quantity: 1,
        unitPrice: 29.99
      }
    ],
    total: 1329.98,
    shippingAddress: {
      street: "123 Main St",
      city: "New York",
      country: "USA",
      zipCode: "10001"
    },
    createdAt: new Date("2024-03-15"),
    deliveredAt: new Date("2024-03-20")
  },
  {
    userId: db.users.findOne({ username: "bob" })._id,
    orderNumber: "ORD-002",
    status: "shipped",
    items: [
      {
        productId: db.products.findOne({ name: "Programming Book" })._id,
        productName: "Programming Book",
        quantity: 1,
        unitPrice: 49.99
      }
    ],
    total: 49.99,
    shippingAddress: {
      street: "456 Oak Ave",
      city: "Los Angeles",
      country: "USA",
      zipCode: "90001"
    },
    createdAt: new Date("2024-04-10"),
    shippedAt: new Date("2024-04-12")
  },
  {
    userId: db.users.findOne({ username: "charlie" })._id,
    orderNumber: "ORD-003",
    status: "processing",
    items: [
      {
        productId: db.products.findOne({ name: "Mechanical Keyboard" })._id,
        productName: "Mechanical Keyboard",
        quantity: 1,
        unitPrice: 89.99
      },
      {
        productId: db.products.findOne({ name: "Wireless Mouse" })._id,
        productName: "Wireless Mouse",
        quantity: 1,
        unitPrice: 29.99
      }
    ],
    total: 119.98,
    shippingAddress: {
      street: "789 Pine Rd",
      city: "Paris",
      country: "France",
      zipCode: "75001"
    },
    createdAt: new Date("2024-04-25")
  }
]);

// Create indexes for better query performance
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ username: 1 }, { unique: true });
db.users.createIndex({ "preferences.theme": 1 });
db.users.createIndex({ isActive: 1 });

db.products.createIndex({ category: 1 });
db.products.createIndex({ price: 1 });
db.products.createIndex({ tags: 1 });
db.products.createIndex({ "specs.cpu": 1 });

db.orders.createIndex({ userId: 1 });
db.orders.createIndex({ status: 1 });
db.orders.createIndex({ orderNumber: 1 }, { unique: true });
db.orders.createIndex({ createdAt: -1 });

// Create a collection with various data types for testing
db.datatypes.insertMany([
  {
    stringField: "Hello World",
    numberField: 42,
    floatField: 3.14159,
    booleanField: true,
    nullField: null,
    dateField: new Date(),
    objectIdField: new ObjectId(),
    arrayField: [1, 2, 3, 4, 5],
    objectField: {
      nested: "value",
      deep: {
        deeper: "value"
      }
    },
    binaryField: BinData(0, "SGVsbG8gV29ybGQ="),
    regexField: /^test/i,
    timestampField: new Timestamp(1714500000, 1)
  }
]);

print("MongoDB seed data inserted successfully!");
