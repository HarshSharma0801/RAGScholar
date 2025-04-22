"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Layout from "@/components/Layout";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { fetchPaperById, PaperById } from "@/services/paperService";
import { format } from "date-fns";
import { ExternalLink, FileText } from "lucide-react";
import SearchBarPaper from '@/components/SearchBarPaper';

export default function PaperPage() {
  const params = useParams();
  const paperId = params.id as string;
  const [paper, setPaper] = useState<PaperById | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchPaper = async () => {
      try {
        setLoading(true);
        const response = await fetchPaperById(paperId);
        if (response) {
          setPaper(response);
        }
      } catch (err) {
        console.error("Failed to fetch paper:", err);
        setError("Failed to load paper");
      } finally {
        setLoading(false);
      }
    };

    fetchPaper();
  }, [paperId]);

  if (loading) {
    return (
      <Layout>
        <div className="space-y-4 animate-pulse max-w-4xl mx-auto">
          <div className="h-8 bg-muted rounded w-3/4"></div>
          <div className="h-4 bg-muted rounded w-1/2"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </Layout>
    );
  }

  if (error) {
    return (
      <Layout>
        <div className="text-center py-10 max-w-4xl mx-auto">
          <h2 className="text-2xl font-bold">Error Loading Paper</h2>
          <p className="text-muted-foreground mt-2">{error}</p>
          <Button className="mt-4" onClick={() => window.history.back()}>
            Go Back
          </Button>
        </div>
      </Layout>
    );
  }

  if (!paper) {
    return (
      <Layout>
        <div className="text-center py-10 max-w-4xl mx-auto">
          <h2 className="text-2xl font-bold">Paper Not Found</h2>
          <p className="text-muted-foreground mt-2">
            The requested paper could not be found.
          </p>
          <Button className="mt-4" onClick={() => window.history.back()}>
            Go Back
          </Button>
        </div>
      </Layout>
    );
  }

  const publishedDate = format(new Date(paper.published), "MMMM d, yyyy");
  const updatedDate = format(new Date(paper.updated), "MMMM d, yyyy");
  const arxivId = paper.id.split("/").pop();

  return (
    <Layout>
      <SearchBarPaper paperContext={paper.title} />
      <div className="space-y-6 max-w-4xl mx-auto">
        <Card className="border-none shadow-none">
          <CardHeader className="pb-2">
            <CardTitle className="text-2xl font-bold">{paper.title}</CardTitle>
              <div className="flex flex-wrap gap-x-2">
                {paper.authors.map((author, index) => (
                  <span key={index}>
                    {author.Name}
                    {index < paper.authors.length - 1 ? "," : ""}
                  </span>
                ))}
              </div>
              <div className="text-sm text-muted-foreground">
                Published: {publishedDate} • Updated: {updatedDate}
              </div>
              {arxivId && (
                <div className="text-sm text-muted-foreground">
                  arXiv: {arxivId}
                </div>
              )}
          </CardHeader>

          <CardContent className="space-y-4">
            {/* Abstract - Fixed section */}
            <section>
              <h3 className="text-lg font-semibold mb-2">Abstract</h3>
              <div className="whitespace-pre-line text-muted-foreground">
                {paper.summary}
              </div>
            </section>

            {/* Links - Highlighted Section */}
            <section>
              <h3 className="text-lg font-semibold mb-3">Download</h3>
              <div className="flex flex-col sm:flex-row gap-3">
                {paper.links.map((link, index) => (
                  <a
                    key={index}
                    href={link.Href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 px-4 py-2 border rounded-md hover:bg-accent transition-colors"
                  >
                    {link.Type === "application/pdf" ? (
                      <>
                        <FileText className="h-4 w-4 text-red-500" />
                        <span>PDF</span>
                      </>
                    ) : (
                      <>
                        <ExternalLink className="h-4 w-4 text-blue-500" />
                        <span>arXiv Page</span>
                      </>
                    )}
                  </a>
                ))}
              </div>
            </section>

            {/* Metadata */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {paper.categories && paper.categories?.length > 0 && (
                <section>
                  <h3 className="text-lg font-semibold mb-2">Categories</h3>
                  <div className="flex flex-wrap gap-2">
                    {paper.categories.map((category, index) => (
                      <span
                        key={index}
                        className="px-2 py-1 bg-secondary text-secondary-foreground rounded-md text-sm"
                      >
                        {category}
                      </span>
                    ))}
                  </div>
                </section>
              )}

              {paper.journalRef && (
                <section>
                  <h3 className="text-lg font-semibold mb-2">
                    Journal Reference
                  </h3>
                  <div className="text-muted-foreground">
                    {paper.journalRef}
                  </div>
                </section>
              )}

              {paper.comment && (
                <section className="md:col-span-2">
                  <h3 className="text-lg font-semibold mb-2">Comments</h3>
                  <div className="text-muted-foreground whitespace-pre-line">
                    {paper.comment}
                  </div>
                </section>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </Layout>
  );
}
