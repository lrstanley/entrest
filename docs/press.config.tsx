import { defineConfig } from "fumapress";
import { fumadocsMdx } from "fumapress/adapters/mdx";
import { metaSchema, pageSchema } from "fumapress/adapters/mdx/schema";
import { createDocsLayoutPage } from "fumapress/layouts/docs";
import { createHomeLayout } from "fumapress/layouts/home";
import { llmsPlugin } from "fumapress/plugins/llms.txt";
import { flexsearchPlugin } from "fumapress/plugins/flexsearch";
import { defineDocs } from "fumadocs-mdx/macro";
import favicon from "./src/assets/favicon.png";
import { Logo } from "./src/components/logo";
import { getMdxComponents } from "./src/components/mdx";

const docs = defineDocs({
    dir: "content",
    docs: {
        async: true,
        schema: pageSchema,
        lastModified: true,
        postprocess: {
            includeProcessedMarkdown: { output: "function" },
        },
    },
    meta: {
        schema: metaSchema,
    },
});

const config = defineConfig({
    content: docs.toFumadocsSource(),
    mode: "static",
    site: {
        name: "entrest",
        baseUrl: "https://lrstanley.github.io/entrest",
        git: {
            user: "lrstanley",
            repo: "entrest",
            branch: "master",
            rootDir: "..",
        },
    },
    defaultLayoutProps: {
        nav: {
            title: <Logo />,
        },
        githubUrl: "https://github.com/lrstanley/entrest",
        links: [
            {
                text: "pkg.go.dev",
                url: "https://pkg.go.dev/github.com/lrstanley/entrest",
                external: true,
            },
            { text: "EntGo", url: "https://entgo.io/", external: true },
            { text: "Discord", url: "https://liam.sh/chat", external: true },
        ],
    },
    renderPage: createDocsLayoutPage({
        render(page) {
            // the page already ends with "Next Steps" cards, a prev/next footer would just repeat them.
            if (page.url === "/guides/getting-started")
                return { pageProps: { footer: { enabled: false } } };
            return {};
        },
    }),
    meta: {
        root() {
            return (
                <>
                    <meta name="author" content="Liam Stanley" />
                    <meta name="copyright" content="© Liam Stanley" />
                    <meta name="darkreader-lock" />
                    <link rel="icon" href={favicon} />
                    <link rel="preconnect" href="https://fonts.googleapis.com" />
                    <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="" />
                    <link
                        href="https://fonts.googleapis.com/css2?family=Geist:wght@100..900&family=JetBrains+Mono:ital,wght@0,100..800;1,100..800&display=swap"
                        rel="stylesheet"
                    />
                </>
            );
        },
    },
})
    .adapters(fumadocsMdx({ getMdxComponents }))
    .plugins(llmsPlugin({ routes: "all" }), flexsearchPlugin());

export default config;
export const HomeLayout = createHomeLayout<typeof config.$context>();
