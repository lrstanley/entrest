import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import starlightLinksValidator from "starlight-links-validator";
import { rehypeHeadingIds, unified } from "@astrojs/markdown-remark";
import rehypeAutolinkHeadings from "rehype-autolink-headings";

process.env.ASTRO_TELEMETRY_DISABLED = "1";

// https://astro.build/config
export default defineConfig({
    site: "https://lrstanley.github.io",
    base: "/entrest",
    srcDir: "./",
    trailingSlash: "always",
    contentDir: "./content",
    integrations: [
        starlight({
            title: "Ent Rest Extension",
            logo: {
                light: "/assets/images/logo-light.webp",
                dark: "/assets/images/logo-dark.webp",
                replacesTitle: true,
            },
            favicon: "/favicon.png",
            social: [
                {
                    icon: "github",
                    label: "GitHub",
                    href: "https://github.com/lrstanley/entrest",
                },
                {
                    icon: "discord",
                    label: "Discord",
                    href: "https://liam.sh/chat",
                },
            ],
            lastUpdated: true,
            editLink: {
                baseUrl: "https://github.com/lrstanley/entrest/edit/master/docs/",
            },
            customCss: ["./assets/main.css"],
            head: [
                {
                    tag: "meta",
                    attrs: {
                        name: "author",
                        content: "Liam Stanley",
                    },
                },
                {
                    tag: "meta",
                    attrs: {
                        name: "copyright",
                        content: "© Liam Stanley",
                    },
                },
                {
                    tag: "meta",
                    attrs: {
                        name: "darkreader-lock",
                    },
                },
            ],
            sidebar: [
                {
                    label: "Guides",
                    items: [{ autogenerate: { directory: "guides" } }],
                },
                {
                    label: "Generating OpenAPI Specs",
                    items: [{ autogenerate: { directory: "openapi-specs" } }],
                },
                {
                    label: "HTTP Handler",
                    items: [{ autogenerate: { directory: "http-handler" } }],
                },
                {
                    label: "Resources",
                    items: [{ autogenerate: { directory: "resources" } }],
                },
                {
                    label: "Resources",
                    badge: {
                        text: "external",
                        variant: "danger",
                    },
                    items: [
                        {
                            label: "Contributing",
                            link: "https://github.com/lrstanley/entrest/blob/master/.github/CONTRIBUTING.md",
                            attrs: {
                                target: "_blank",
                            },
                        },
                        {
                            label: "GitHub Project",
                            link: "https://github.com/lrstanley/entrest",
                            attrs: {
                                target: "_blank",
                            },
                        },
                        {
                            label: "pkg.go.dev docs",
                            link: "https://pkg.go.dev/github.com/lrstanley/entrest",
                            attrs: {
                                target: "_blank",
                            },
                        },
                        {
                            label: "EntGo Documentation",
                            link: "https://entgo.io/",
                            attrs: {
                                target: "_blank",
                            },
                        },
                    ],
                },
            ],
            plugins: [
                starlightLinksValidator({
                    errorOnLocalLinks: false,
                    errorOnRelativeLinks: false,
                }),
            ],
        }),
    ],
    markdown: {
        processor: unified({
            gfm: true,
            // SmartyPants converts '--' into en-dash, breaking alignment.
            smartypants: false,
            rehypePlugins: [rehypeHeadingIds, [rehypeAutolinkHeadings, { behavior: "wrap" }]],
        }),
    },
});
