"use client";

import { useSearchParams } from "next/navigation";
import Layout from "@/components/Layout";
import PaperCard from "@/components/PaperCard";
import { analyzePaper } from "@/services/paperService";
import { useQuery } from "@tanstack/react-query";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Paper } from "@/types/paper";

export default function SearchPage() {
  const searchParams = useSearchParams();
  const query = searchParams.get("q") || "";
  const text = searchParams.get("t") || "";
  const context = searchParams.get("c") || "";

  const { data, isLoading } = useQuery({
    queryKey: ["searchPapers", query],
    queryFn: () =>
      analyzePaper({
        searchQuery: query,
        selectedText: text,
        paperContext: context,
      }),
    enabled: !!query,
  });

  const explanation = data?.explanation || "No explanation available.";

  return (
    <Layout>
      <div className="space-y-8">
        {/* Explanation Section with Typewriter Effect */}
        <div className="bg-gradient-to-br from-blue-50 to-purple-50 rounded-xl p-6 shadow-sm border border-gray-100">
          <h1 className="text-2xl font-bold tracking-tight mb-4 text-gray-800">
            Explanation
          </h1>

          {/* Static Markdown Render for Reference */}
          {!isLoading && data?.explanation && (
            <div className="mt-6 border-t pt-6">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  code({ className, children, ...props }) {
                    return (
                      <code className={className} {...props}>
                        {children}
                      </code>
                    );
                  },
                  h2: ({ children, ...props }) => (
                    <h2
                      className="text-xl font-semibold mt-4 mb-2 text-gray-800"
                      {...props}
                    >
                      {children}
                    </h2>
                  ),
                  p: ({ children, ...props }) => (
                    <p className="mb-4 leading-relaxed" {...props}>
                      {children}
                    </p>
                  ),
                  ul: ({ children, ...props }) => (
                    <ul className="list-disc pl-5 mb-4 space-y-1" {...props}>
                      {children}
                    </ul>
                  ),
                  li: ({ children, ...props }) => (
                    <li className="mb-1" {...props}>
                      {children}
                    </li>
                  ),
                  strong: ({ children, ...props }) => (
                    <strong className="font-semibold text-gray-800" {...props}>
                      {children}
                    </strong>
                  ),
                  em: ({ children, ...props }) => (
                    <em className="italic" {...props}>
                      {children}
                    </em>
                  ),
                  blockquote: ({ children, ...props }) => (
                    <blockquote
                      className="border-l-4 border-blue-300 pl-4 italic text-gray-600"
                      {...props}
                    >
                      {children}
                    </blockquote>
                  ),
                }}
              >
                {explanation}
              </ReactMarkdown>
            </div>
          )}
        </div>

        {/* Papers Section */}
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="h-64 rounded-lg bg-muted animate-pulse" />
            ))}
          </div>
        ) : data?.relatedPapers?.length > 0 ? (
          <>
            <div className="mt-8">
              <h1 className="text-2xl font-bold tracking-tight text-gray-800">
                Related Research Papers
              </h1>
              <p className="text-muted-foreground mt-2">
                Found {data.relatedPapers.length} relevant papers
              </p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {data.relatedPapers.map((paper: { paper: Paper }) => (
                <PaperCard key={paper.paper.id} paper={paper.paper} />
              ))}
            </div>
          </>
        ) : (
          <div className="text-center py-12 bg-gray-50 rounded-xl">
            <h3 className="text-lg font-medium text-gray-800">
              No results found
            </h3>
            <p className="text-muted-foreground mt-2">
              Try searching with different keywords or check your spelling.
            </p>
          </div>
        )}
      </div>
    </Layout>
  );
}
