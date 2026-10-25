const links = [
    { href: "/llms.txt", label: "llms.txt" },
    { href: "/llms-full.txt", label: "llms-full.txt" },
];

export function LlmsLinks({ className = "" }: { className?: string }) {
    return (
        <p className={`flex flex-wrap items-center gap-x-2 text-xs text-fd-muted-foreground ${className}`}>
            <span>For LLMs:</span>
            {links.map(({ href, label }, i) => (
                <span key={href} className="contents">
                    {i > 0 && <span aria-hidden="true">·</span>}
                    <a href={href} className="font-mono underline-offset-4 hover:text-fd-foreground hover:underline">
                        {label}
                    </a>
                </span>
            ))}
        </p>
    );
}
