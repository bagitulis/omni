const bcrypt = require("bcryptjs");

// Generate new hash
const newHash = bcrypt.hashSync("password123", 10);
console.log("New hash:", newHash);
console.log("Length:", newHash.length);

// Verify
const testHash = "$2b$10$on9zfFqJ18gPPGxJT8/Idu0kW5LUPUCpVbitONn7sT5oY1TW5RSiW";
const match1 = bcrypt.compareSync("password123", testHash);
console.log("Match with test hash:", match1);

const match2 = bcrypt.compareSync("password123", newHash);
console.log("Match with new hash:", match2);
