import { Children, isValidElement, type ComponentProps, type ReactNode } from "react";
import defaultMdxComponents, { createRelativeLink } from "fumadocs-ui/mdx";
import { Callout as FDCallout } from "fumadocs-ui/components/callout";
import { Card as FDCard, Cards as FDCards } from "fumadocs-ui/components/card";
import { Step as FDStep, Steps as FDSteps } from "fumadocs-ui/components/steps";
import { asMarkdown, md } from "fumapress/markdown";
import type { MDXComponents } from "mdx/types";
import { PeleError, render } from "pele";

// Components below call asMarkdown() so the generated `/<page>.md`, llms.txt and
// llms-full.txt contain plain markdown rather than raw JSX.

const calloutLabels: Record<string, string> = {
    info: "Note",
    idea: "Tip",
    warn: "Warning",
    error: "Danger",
};

function Callout({ title, children, type = "info", ...props }: ComponentProps<typeof FDCallout>) {
    if (asMarkdown())
        return md.linePrefix("> ")`**${title ?? calloutLabels[type] ?? "Note"}**\n${children}`;
    return (
        <FDCallout title={title} type={type} {...props}>
            {children}
        </FDCallout>
    );
}

function Cards({ children, ...props }: ComponentProps<typeof FDCards>) {
    if (asMarkdown()) return md`${children}`;
    return <FDCards {...props}>{children}</FDCards>;
}

function Card({ title, description, href, children, ...props }: ComponentProps<typeof FDCard>) {
    if (asMarkdown()) {
        const label = href ? `[${title}](${href})` : `**${title}**`;
        return md`- ${label}${description ? `: ${description}` : ""}\n`;
    }
    return (
        <FDCard title={title} description={description} href={href} {...props}>
            {children}
        </FDCard>
    );
}

function Steps({ children }: { children: ReactNode }) {
    if (asMarkdown()) {
        return (
            <ol>
                {Children.toArray(children)
                    .filter(isValidElement)
                    .map((child, i) => (
                        <li key={i}>{child}</li>
                    ))}
            </ol>
        );
    }
    return <FDSteps>{children}</FDSteps>;
}

function Step({ children }: { children: ReactNode }) {
    if (asMarkdown()) return <>{children}</>;
    return <FDStep>{children}</FDStep>;
}

/** Rendered to inline SVG at build time with Pele, so no mermaid runtime is shipped to the browser. */
function Mermaid({ chart }: { chart: string }) {
    if (asMarkdown()) return md`\`\`\`mermaid\n${chart}\n\`\`\`\n`;

    try {
        const { svg } = render(chart, {
            idPrefix: `pele-${hash(chart)}`,
            fontFamily: "Geist, sans-serif",
        });
        return (
            <div
                className="my-6 flex justify-center overflow-x-auto"
                dangerouslySetInnerHTML={{ __html: svg }}
            />
        );
    } catch (error) {
        if (!(error instanceof PeleError)) throw error;
        // fail the build instead of silently shipping a broken diagram.
        throw new Error(`invalid mermaid diagram: ${error.message}\n${error.snippet}`, {
            cause: error,
        });
    }
}

function hash(s: string) {
    let h = 0;
    for (const c of s) h = (Math.imul(31, h) + c.charCodeAt(0)) | 0;
    return (h >>> 0).toString(36);
}

const usageStyles = {
    schema: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300",
    edge: "bg-red-500/15 text-red-700 dark:text-red-300",
    field: "bg-sky-500/15 text-sky-700 dark:text-sky-300",
} as const;

/** Marks where an annotation may be applied: schema, edge and/or field. */
function Usage({ types }: { types: (keyof typeof usageStyles)[] }) {
    if (asMarkdown()) return md`${types.join(", ")}`;

    return (
        <span className="inline-flex gap-1 align-middle">
            {types.map((t) => (
                <span
                    key={t}
                    className={`rounded px-1.5 py-0.5 text-xs font-medium ${usageStyles[t]}`}
                >
                    {t}
                </span>
            ))}
        </span>
    );
}

export const getMdxComponents = async function (this: any, page: any): Promise<MDXComponents> {
    return {
        ...defaultMdxComponents,
        a: createRelativeLink(await this.getLoader(), page),
        Callout,
        Card,
        Cards,
        Step,
        Steps,
        Mermaid,
        Usage,
    };
};
