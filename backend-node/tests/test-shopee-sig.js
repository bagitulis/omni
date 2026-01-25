const crypto = require("crypto");

// From the logs - actual webhook data
const rawBody =
  '{"msg_id":"1ed4ce6347d5e9ab48f7aeba3cd64800","data":{"ordersn":"260107EHFXHD2W","forder_id":"5869348275461129055","package_number":"OFG221499822241630","tracking_no":"SPXID068120696021"},"shop_id":530635055,"code":4,"timestamp":1767833934}';
const expectedSig =
  "18762f3f62bae045e0c2e6419142676dfcba12bf98cc527be9d58bbe36c3ec81";

// Stored in database (hex encoded)
const hexKey =
  "544c48794b6251566b614e6246666d616b76447676674a54794677574d786d58";
// Decoded key
const decodedKey = Buffer.from(hexKey, "hex").toString("utf8");

console.log("Decoded key:", decodedKey);
console.log("");

console.log("Using HEX key as-is:");
const sig1 = crypto.createHmac("sha256", hexKey).update(rawBody).digest("hex");
console.log("  Computed:", sig1);
console.log("  Expected:", expectedSig);
console.log("  Match:", sig1 === expectedSig);

console.log("");
console.log("Using DECODED key:");
const sig2 = crypto
  .createHmac("sha256", decodedKey)
  .update(rawBody)
  .digest("hex");
console.log("  Computed:", sig2);
console.log("  Expected:", expectedSig);
console.log("  Match:", sig2 === expectedSig);
