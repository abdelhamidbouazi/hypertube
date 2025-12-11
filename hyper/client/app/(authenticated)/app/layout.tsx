import { LayoutWrapper } from "@/components/LayoutWrapper";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return <LayoutWrapper>{children}</LayoutWrapper>;
}
