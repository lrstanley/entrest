import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import press from "fumapress/vite";
import { fumadocsMdx } from "fumadocs-mdx/vite";

// The adapter is pinned to node: fumapress otherwise switches to its Cloudflare adapter in Workers Builds
// (WORKERS_CI), which rewrites images to /cdn-cgi/image/ transformation URLs that 404 on workers.dev. The
// site is static either way; public/_headers keeps the asset caching that adapter would add.
export default defineConfig({
  plugins: [press({ adapter: "waku/adapters/node" }), fumadocsMdx(), tailwindcss()],
});
