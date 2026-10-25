// Post-build checks and fixups for the static site in dist/public.
//
//  1. Markdown outputs (llms.txt, llms-full.txt, /<page>.md) are generated with root-relative links
//     (e.g. /guides/getting-started), so rewrite them to absolute URLs.
//  2. Point the llms.txt page links at the markdown version of each page (/<page>.md, /index.md for the root),
//     per https://llmstxt.org, and check those files exist.
//  3. Validate every internal link and #fragment in the generated HTML.
import { existsSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const SITE = "https://entrest.dev";
const OUT = join(import.meta.dirname, "..", "dist", "public");

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((f) => {
    const p = join(dir, f);
    return statSync(p).isDirectory() ? walk(p) : [p];
  });
}

const files = walk(OUT);
const errors: string[] = [];

for (const file of files.filter((f) => f.endsWith(".md") || /llms(-full)?\.txt$/.test(f))) {
  const src = readFileSync(file, "utf8");
  const out = src
    .replace(/\]\(\/(?!\/)/g, `](${SITE}/`) // markdown links
    .replace(/^(# .+) \(\/([^)]*)\)$/gm, `$1 (${SITE}/$2)`); // llms-full.txt page headings
  if (out !== src) writeFileSync(file, out);
}

// llms.txt links should point at LLM-friendly content, so use the generated markdown version of each page.
const llmsPath = join(OUT, "llms.txt");
writeFileSync(
  llmsPath,
  readFileSync(llmsPath, "utf8").replace(
    new RegExp(`\\]\\(${SITE.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}(/[^)#]*)?\\)`, "g"),
    (_, path = "") => {
      const md = (path.replace(/\/$/, "") || "/index") + ".md";
      if (!existsSync(join(OUT, md))) errors.push(`llms.txt: no markdown file for ${SITE}${path} (${md})`);
      return `](${SITE}${md})`;
    },
  ),
);

// The generated llms.txt index is titled "Docs" with no summary, give it the project name, description and
// the external references that the previous Starlight setup provided.
const llms = readFileSync(llmsPath, "utf8");
if (llms.startsWith("# Docs")) {
  writeFileSync(
    llmsPath,
    llms
      .replace(
        "# Docs",
        "# entrest\n\n> entrest is an EntGo extension for generating compliant OpenAPI specs and HTTP handler implementations that match those specs.",
      )
      .trimEnd() +
      `

## Optional

- [pkg.go.dev documentation](https://pkg.go.dev/github.com/lrstanley/entrest): Go package API reference on pkg.go.dev
- [EntGo documentation](https://entgo.io/docs/getting-started): official EntGo framework documentation
`,
  );
}

const htmlFiles = files.filter((f) => f.endsWith(".html"));
const ids = new Map<string, Set<string>>();
const idsOf = (file: string) => {
  let set = ids.get(file);
  if (!set) {
    set = new Set(
      [...readFileSync(file, "utf8").matchAll(/\sid="([^"]+)"/g)].map((m) => m[1] as string),
    );
    ids.set(file, set);
  }
  return set;
};

for (const file of htmlFiles) {
  const page = "/" + file.slice(OUT.length + 1).replace(/(^|\/)index\.html$/, "");
  for (const [, href = ""] of readFileSync(file, "utf8").matchAll(/<a\s[^>]*?href="([^"]*)"/g)) {
    if (/^(https?:|mailto:|tel:|\/\/)/.test(href)) continue;

    const [path = "", fragment] = href.replace(/&amp;/g, "&").split("#");
    let target = file;
    if (path) {
      if (!path.startsWith("/")) {
        errors.push(`${page}: link ${href} is not root-relative`);
        continue;
      }
      const rel = path.replace(/^\//, "").replace(/\/$/, "");
      const candidates = [join(OUT, rel, "index.html"), join(OUT, rel)];
      const found = candidates.find((c) => existsSync(c) && statSync(c).isFile());
      if (!found) {
        errors.push(`${page}: broken link ${href}`);
        continue;
      }
      target = found;
    }
    if (fragment && target.endsWith(".html") && !idsOf(target).has(decodeURIComponent(fragment))) {
      errors.push(`${page}: broken anchor ${href}`);
    }
  }
}

if (errors.length) {
  console.error(
    `\n${errors.length} broken link(s):\n` + [...new Set(errors)].map((e) => `  - ${e}`).join("\n"),
  );
  process.exit(1);
}
console.log(
  `[postbuild] checked links in ${htmlFiles.length} pages, rewrote markdown links to ${SITE}`,
);
