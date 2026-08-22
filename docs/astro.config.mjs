import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import starlightLinksValidator from "starlight-links-validator";
import starlightLlmsTxt from "starlight-llms-txt";
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
            description:
                "entrest is an EntGo extension for generating compliant OpenAPI specs and HTTP handler implementations that match those specs.",
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
                starlightLlmsTxt({
                    projectName: "entrest",
                    description:
                        "entrest is an EntGo extension for generating compliant OpenAPI specs and HTTP handler implementations that match those specs.",
                    optionalLinks: [
                        {
                            label: "pkg.go.dev documentation",
                            url: "https://pkg.go.dev/github.com/lrstanley/entrest",
                            description: "Go package API reference on pkg.go.dev",
                        },
                        {
                            label: "EntGo documentation",
                            url: "https://entgo.io/docs/getting-started",
                            description: "official EntGo framework documentation",
                        },
                    ],
                    customSets: [
                        {
                            label: "Guides",
                            description: "step-by-step guides for getting started with entrest",
                            paths: ["guides/**"],
                        },
                        {
                            label: "OpenAPI specs",
                            description:
                                "documentation for generating and configuring OpenAPI specs",
                            paths: ["openapi-specs/**"],
                        },
                        {
                            label: "HTTP handler",
                            description:
                                "documentation for the generated HTTP handler and API docs",
                            paths: ["http-handler/**"],
                        },
                        {
                            label: "Resources",
                            description: "architecture, best practices, and troubleshooting",
                            paths: ["resources/**"],
                        },
                    ],
                    promote: ["index*", "guides/getting-started*"],
                }),
                starlightLinksValidator({
                    errorOnLocalLinks: false,
                    errorOnRelativeLinks: false,
                }),
            ],
        }),
    ],
    vite: {
        server: {
            watch: {
                // With srcDir: "./", content-sync writes under .astro/ look like source
                // changes and retrigger Astro 7 route HMR, which logs
                // "Failed to update routes via HMR: TypeError: undefined is not a function".
                ignored: ["**/.astro/**"],
            },
        },
    },
    markdown: {
        processor: unified({
            gfm: true,
            // SmartyPants converts '--' into en-dash, breaking alignment.
            smartypants: false,
            rehypePlugins: [rehypeHeadingIds, [rehypeAutolinkHeadings, { behavior: "wrap" }]],
        }),
    },
});
