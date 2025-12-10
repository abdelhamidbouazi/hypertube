export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <section className="flex flex-col items-center justify-start gap-4 pb-8 w-full">
      <div className="w-full text-center justify-start">{children}</div>
    </section>
  );
}
