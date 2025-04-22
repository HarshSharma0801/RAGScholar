import React from "react";
import SearchBar from "./SearchBar";
import Link from "next/link";

interface LayoutProps {
  children: React.ReactNode;
  paperContext?: string;
}

const Layout: React.FC<LayoutProps> = ({ children, paperContext }) => {
  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-10 border-b bg-background/95 backdrop-blur">
        <div className="container mx-auto py-4">
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <Link href="/" className="text-xl font-bold">
                RAGScholar
              </Link>
            </div>
          </div>
        </div>
      </header>
      <main className="container mx-auto py-6">{children}</main>
    </div>
  );
};

export default Layout;
