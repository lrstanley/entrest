import { defineCollection, z } from "astro:content";
import { docsSchema } from "@astrojs/starlight/schema";

export const collections = {
  docs: defineCollection({
    schema: docsSchema({
      extend: z.object({
        // global banner thanks to: https://hideoo.dev/notes/starlight-sitewide-banner
        banner: z.object({ content: z.string() }).default({
          content: `Documentation &amp; entrest itself are a work in progress (expect breaking changes). Check out the <a href="https://github.com/lrstanley/entrest" target="_blank">GitHub Project</a> to contribute.`,
        }),
      }),
    }),
  }),
};
