import type { Root } from "mdast";
import type { VFile } from "vfile";
import { visit } from "unist-util-visit";
import { replaceSource } from "fumadocs-core/mdx-plugins/stringifier";

/** Normalizes code pulled in from `_examples/` so it reads like a standalone project. */
export default function remarkCleanExamples() {
  return (tree: Root, file: VFile) => {
    visit(tree, "code", (node) => {
      const cleaned = node.value
        .replace(/\t/g, "    ")
        .replace(
          /github.com\/lrstanley\/entrest\/_examples\/[^/]+/g,
          "github.com/example/my-project",
        );
      if (cleaned === node.value) return;
      node.value = cleaned;

      let fence = "```";
      while (cleaned.includes(fence)) fence += "`";
      replaceSource(
        file,
        node,
        `${fence}${node.lang ?? ""}${node.meta ? ` ${node.meta}` : ""}\n${cleaned}\n${fence}`,
      );
    });
  };
}
