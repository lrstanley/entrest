export function Logo({ className = "text-lg" }: { className?: string }) {
    return (
        <span className={`font-mono font-bold tracking-tight ${className}`}>
            <span className="text-fd-primary">ent.</span>
            <span>rest</span>
        </span>
    );
}
