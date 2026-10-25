import { Link } from "fumapress/client";
import { asMarkdown, md } from "fumapress/markdown";
import { ArrowDownUp, ArrowRight, BookOpen, FileJson, Filter, Rocket, Server } from "lucide-react";
import hero from "../assets/hero.png";
import { Logo } from "../components/logo";

const features = [
    {
        icon: FileJson,
        title: "Compliant OpenAPI Spec",
        body: "Generated OpenAPI spec for your EntGo schema.",
    },
    {
        icon: Server,
        title: "HTTP Handler",
        body: "Fully-functional HTTP handler implementation that matches the OpenAPI spec.",
    },
    {
        icon: BookOpen,
        title: "Supports pagination",
        body: "Automatic pagination (where applicable).",
    },
    {
        icon: Filter,
        title: "Supports filtering",
        body: "Advanced filtering (using query parameters, AND/OR predicates, etc).",
    },
    {
        icon: Rocket,
        title: "Supports eager-loading",
        body: "Supports eager-loading edges, so you don't have to make additional calls unnecessarily.",
    },
    { icon: ArrowDownUp, title: "Supports sorting", body: "Supports various forms of sorting." },
];

const tagline =
    "ent.rest is an EntGo extension for generating compliant OpenAPI specs and an HTTP handler implementation that matches that spec.";

export default function Page() {
    if (asMarkdown()) {
        return md`
# ent.rest

${tagline}

${features.map((f) => `- **${f.title}**: ${f.body}`).join("\n")}

- [Getting Started](/guides/getting-started)
- [Source](https://github.com/lrstanley/entrest)
`;
    }

    return (
        <>
            <title>ent.rest - Generated OpenAPI specs and HTTP handlers for EntGo</title>
            <meta
                name="description"
                content="Generate compliant OpenAPI specs and HTTP handlers from your EntGo schemas with entrest."
            />
            <main className="mx-auto flex w-full max-w-6xl flex-col gap-16 px-6 py-16 md:py-24">
                <section className="flex flex-col items-center gap-8 text-center">
                    <h1>
                        <Logo className="text-5xl md:text-7xl" />
                    </h1>
                    <p className="max-w-2xl text-lg text-fd-muted-foreground">
                        <code className="text-fd-primary">ent.rest</code> is an{" "}
                        <a
                            href="https://entgo.io/"
                            target="_blank"
                            rel="noreferrer"
                            className="underline underline-offset-4"
                        >
                            EntGo
                        </a>{" "}
                        extension for generating compliant OpenAPI specs and an HTTP handler
                        implementation that matches that spec.
                    </p>
                    <div className="flex flex-wrap justify-center gap-3">
                        <Link
                            href="/guides/getting-started"
                            className="inline-flex items-center gap-2 rounded-lg bg-fd-primary px-5 py-2.5 text-sm font-medium text-fd-primary-foreground transition-opacity hover:opacity-90"
                        >
                            Getting Started <ArrowRight className="size-4" />
                        </Link>
                        <a
                            href="https://github.com/lrstanley/entrest"
                            className="inline-flex items-center gap-2 rounded-lg border border-fd-border bg-fd-card px-5 py-2.5 text-sm font-medium transition-colors hover:bg-fd-accent"
                        >
                            View the source
                        </a>
                    </div>
                    <img
                        src={hero}
                        alt="API Docs"
                        className="mt-4 w-full max-w-3xl rounded-xl border border-fd-border shadow-2xl shadow-fd-primary/10"
                    />
                </section>

                <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {features.map(({ icon: Icon, title, body }) => (
                        <div
                            key={title}
                            className="rounded-xl border border-fd-border bg-fd-card p-5"
                        >
                            <Icon className="mb-3 size-5 text-fd-primary" />
                            <h3 className="mb-1 font-semibold">{title}</h3>
                            <p className="text-sm text-fd-muted-foreground">{body}</p>
                        </div>
                    ))}
                </section>
            </main>
        </>
    );
}
