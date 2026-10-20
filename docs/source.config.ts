import { defineConfig } from "fumadocs-mdx/config";
import { rehypeCodeDefaultOptions } from "fumadocs-core/mdx-plugins/rehype-code";
import { remarkMdxMermaid } from "fumadocs-core/mdx-plugins/remark-mdx-mermaid";
import { transformerMetaHighlight, transformerMetaWordHighlight } from "@shikijs/transformers";
import remarkCleanExamples from "./src/lib/remark-clean-examples";

export default defineConfig({
  mdxOptions: {
    remarkPlugins: (v) => [...v, remarkCleanExamples, remarkMdxMermaid],
    rehypeCodeOptions: {
      themes: { light: "catppuccin-latte", dark: "catppuccin-mocha" },
      transformers: [
        ...(rehypeCodeDefaultOptions.transformers ?? []),
        // supports ```go {1,3-5} and ```http /word/ meta strings
        transformerMetaHighlight(),
        transformerMetaWordHighlight(),
      ],
    },
  },
});
