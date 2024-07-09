import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import starlightLinksValidator from "starlight-links-validator";

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
            social: {
                github: "https://github.com/lrstanley/entrest",
                discord: "https://liam.sh/chat",
            },
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
                    autogenerate: {
                        directory: "guides",
                    },
                },
                {
                    label: "Generating OpenAPI Specs",
                    autogenerate: {
                        directory: "openapi-specs",
                    },
                },
                {
                    label: "HTTP Handler",
                    autogenerate: {
                        directory: "http-handler",
                    },
                },
                {
                    label: "Resources",
                    autogenerate: {
                        directory: "resources",
                    },
                },
                {
                    label: "Resources",
                    badge: {
                        text: "external",
                        variant: "danger",
                    },
                    items: [
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
            plugins: [starlightLinksValidator()],
        }),
    ],
    markdown: {
        smartypants: false,
        // SmartyPants converts '--' into en-dash, breaking alignment.
        gfm: true,
    },
});
