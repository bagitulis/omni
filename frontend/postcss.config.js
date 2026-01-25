export default {
  plugins: {
    "@tailwindcss/postcss": {},
    autoprefixer: {
      // Only add prefixes for browsers we need to support
      overrideBrowserslist: ["last 2 versions", "not dead"],
    },
  },
};
