import Link from "next/link";

export default function HomePage() {
  return (
    <div className="flex flex-1 flex-col justify-center text-center">
      <h1 className="mb-4 text-3xl font-bold">warhol</h1>
      <p className="mx-auto max-w-2xl text-fd-muted-foreground">
        A CLI for generating images in a consistent visual style.
      </p>
      <p className="mt-4">
        Start with{" "}
        <Link href="/docs/getting-started" className="font-medium underline">
          Getting Started
        </Link>{" "}
        or browse the{" "}
        <Link href="/docs/cli" className="font-medium underline">
          CLI reference
        </Link>
        .
      </p>
    </div>
  );
}
