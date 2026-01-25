/**
 * Jest Setup File
 * Runs before all tests
 */

// Mock UUID to avoid ESM issues
jest.mock("uuid", () => ({
  v4: jest.fn(() => "00000000-0000-0000-0000-000000000000"),
}));

// Mock googleapis to avoid ESM issues with uuid
jest.mock("googleapis", () => ({
  google: {
    auth: {
      GoogleAuth: jest.fn().mockImplementation(() => ({
        getClient: jest.fn(),
      })),
    },
    sheets: jest.fn(() => ({
      spreadsheets: {
        values: {
          get: jest.fn(),
          update: jest.fn(),
          append: jest.fn(),
          batchUpdate: jest.fn(),
        },
      },
    })),
  },
}));

// Mock environment variables for testing
process.env.NODE_ENV = "test";
process.env.JWT_SECRET = "test-jwt-secret-key-for-testing";
process.env.DATABASE_URL = "file:./test.db";
process.env.DEFAULT_TENANT = "test-tenant";

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(), // Suppress console.log
  debug: jest.fn(), // Suppress console.debug
  info: jest.fn(), // Suppress console.info
  warn: jest.fn(), // Keep warnings visible
  error: jest.fn(), // Keep errors visible
};

// Global test timeout
jest.setTimeout(10000);

// Use fake timers to prevent timer leaks
beforeAll(() => {
  jest.useFakeTimers({ advanceTimers: true });
});

// Global teardown - cleanup all resources after all tests
afterAll(async () => {
  // Clear all pending timers
  jest.clearAllTimers();
  
  // Restore real timers
  jest.useRealTimers();
  
  // Reset all mocks
  jest.resetAllMocks();
  
  // Allow pending microtasks to complete
  await new Promise((resolve) => setImmediate(resolve));
});
