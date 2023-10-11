// SPDX-FileCopyrightText: 2023 froggie <legal@frogg.ie>
//
// SPDX-License-Identifier: OSL-3.0

const { defineConfig } = require("@vue/cli-service")
const path = require("path")

require("dotenv").config({
  path: path.join(__dirname, "..", ".env")
});

module.exports = defineConfig({
  transpileDependencies: true,
  assetsDir: "static",
  devServer: {
    proxy: 'http://localhost:3000',
    port: process.env.DEV_PORT,
  },
})
