const http = require("http");

// First login
const loginData = JSON.stringify({
  username: "yumna",
  password: "password123",
});
const loginOptions = {
  hostname: "localhost",
  port: 3000,
  path: "/api/auth/login",
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Content-Length": loginData.length,
  },
};

console.log("1. Logging in...");
const loginReq = http.request(loginOptions, (res) => {
  let body = "";
  res.on("data", (d) => (body += d));
  res.on("end", () => {
    console.log("Login Status:", res.statusCode);
    const loginResult = JSON.parse(body);

    if (res.statusCode !== 200) {
      console.log("Login failed:", body);
      return;
    }

    const token = loginResult.token;
    const tenantId = loginResult.tenantId;

    console.log(
      "Token obtained (first 50 chars):",
      token.substring(0, 50) + "...",
    );
    console.log("Tenant ID:", tenantId);

    // Test token-status
    console.log("\n2. Testing /api/token-status...");
    const statusOptions = {
      hostname: "localhost",
      port: 3000,
      path: "/api/token-status",
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
        "x-tenant-id": tenantId,
      },
    };

    const statusReq = http.request(statusOptions, (res2) => {
      let body2 = "";
      res2.on("data", (d) => (body2 += d));
      res2.on("end", () => {
        console.log("Token Status Code:", res2.statusCode);
        console.log("Token Status Body:", body2);

        // Test tokens/status
        console.log("\n3. Testing /api/tokens/status...");
        const tokensOptions = {
          hostname: "localhost",
          port: 3000,
          path: "/api/tokens/status",
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
            "x-tenant-id": tenantId,
          },
        };

        const tokensReq = http.request(tokensOptions, (res3) => {
          let body3 = "";
          res3.on("data", (d) => (body3 += d));
          res3.on("end", () => {
            console.log("Tokens Status Code:", res3.statusCode);
            console.log("Tokens Status Body:", body3);

            // Test tokens/refresh/shopee
            console.log("\n4. Testing /api/tokens/refresh/shopee...");
            const refreshOptions = {
              hostname: "localhost",
              port: 3000,
              path: "/api/tokens/refresh/shopee",
              method: "POST",
              headers: {
                Authorization: `Bearer ${token}`,
                "x-tenant-id": tenantId,
              },
            };

            const refreshReq = http.request(refreshOptions, (res4) => {
              let body4 = "";
              res4.on("data", (d) => (body4 += d));
              res4.on("end", () => {
                console.log("Refresh Status Code:", res4.statusCode);
                console.log("Refresh Body:", body4);
              });
            });
            refreshReq.end();
          });
        });
        tokensReq.end();
      });
    });
    statusReq.end();
  });
});
loginReq.write(loginData);
loginReq.end();
