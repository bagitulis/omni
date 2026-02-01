function calculateDiscount(price, discountPercent) {
  // Validate inputs
  if (typeof price !== "number" || typeof discountPercent !== "number") {
    throw new TypeError("Both price and discountPercent must be numbers");
  }

  if (price < 0) {
    throw new Error("Price cannot be negative");
  }

  if (discountPercent < 0) {
    throw new Error("Discount percent cannot be negative");
  }

  if (discountPercent > 100) {
    return price;
  }

  const discount = (price * discountPercent) / 100;
  return price - discount;
}

// Test cases
console.log(calculateDiscount(100, 0)); // 100 ✓
console.log(calculateDiscount(100, 10)); // 90 ✓
console.log(calculateDiscount(100, 100)); // 0 ✓
console.log(calculateDiscount(100, 150)); // 100 ✓

try {
  calculateDiscount(-50, 10); // throws error ✓
} catch (e) {
  console.log("Caught:", e.message);
}

try {
  calculateDiscount(100, -10); // throws error ✓
} catch (e) {
  console.log("Caught:", e.message);
}
