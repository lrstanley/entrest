import logoDark from "../assets/logo-dark.webp";
import logoLight from "../assets/logo-light.webp";

export function Logo({ className = "h-6" }: { className?: string }) {
    return (
        <>
            <img src={logoLight} alt="entrest" className={`${className} w-auto dark:hidden`} />
            <img src={logoDark} alt="entrest" className={`${className} hidden w-auto dark:block`} />
        </>
    );
}
