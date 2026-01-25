const JWT_TOKEN =
  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIyNTE1Mjg4ZC01MzczLTQzNmQtODU2NC03NmZjNDE3MDEzNjAiLCJ0ZW5hbnRJZCI6Inl1bW5hX2JlcnRpZ2FtYXJ0Iiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzY5MjU3NzA0LCJpYXQiOjE3NjkxNzEzMDR9.y3aSAGo_tVYypihiWSIcVw1yMNrDSMZqzLUaRgsd7V4";

// Use nginx container to make requests
const { exec } = require("child_process");

function testAPI(endpoint, description, method = "GET") {
  return new Promise((resolve, reject) => {
    let cmd;
    if (method === "POST") {
      cmd = `docker exec omni-backend wget -qO- --post-data="" --header="Authorization: Bearer ${JWT_TOKEN}" --header="Content-Type: application/json" "http://localhost:3000${endpoint}"`;
    } else {
      cmd = `docker exec omni-backend wget -qO- --header="Authorization: Bearer ${JWT_TOKEN}" "http://localhost:3000${endpoint}"`;
    }
    exec(cmd, { maxBuffer: 10 * 1024 * 1024 }, (error, stdout, stderr) => {
      console.log(`\n=== ${description} ===`);
      console.log(`${method} ${endpoint}`);
      if (error) {
        console.log("Error:", stderr || error.message);
        resolve(null);
        return;
      }
      try {
        const data = JSON.parse(stdout);
        // Show count and first few items
        if (data.data && data.data.processed) {
          console.log("Sync Results:");
          for (const [platform, result] of Object.entries(
            data.data.processed.data || {},
          )) {
            console.log(
              `  ${platform}: success=${result.success}, count=${result.count}`,
            );
          }
        } else if (data.data && Array.isArray(data.data)) {
          console.log(`Total: ${data.total || data.data.length} orders`);
          console.log("First 2 orders:");
          data.data.slice(0, 2).forEach((o, i) => {
            console.log(
              `  ${i + 1}. ${o.order_no || o.orderSn || o.order_sn} - ${o.platform} - ${o.status || o.order_status}`,
            );
          });
        } else {
          console.log(
            "Response:",
            JSON.stringify(data, null, 2).substring(0, 800),
          );
        }
        resolve(data);
      } catch (e) {
        console.log("Raw Response:", stdout.substring(0, 500));
        resolve(stdout);
      }
    });
  });
}

async function main() {
  console.log("Testing Order Manager API Endpoints...\n");

  // Test processed orders (GET)
  await testAPI(
    "/api/orders/processed?page=1&limit=10",
    "GET Processed Orders",
  );

  // Test unprocess orders (GET)
  await testAPI(
    "/api/orders/unprocess?page=1&limit=10",
    "GET Unprocess Orders",
  );

  // Test unpaid orders (GET)
  await testAPI("/api/orders/unpaid?page=1&limit=10", "GET Unpaid Orders");

  // Test sync-all (POST)
  await testAPI("/api/orders/sync-all", "Sync All Orders", "POST");

  console.log("\n=== TEST COMPLETE ===");
}

main();
