'use client';

import Layout from '@/components/Layout';
import PaperCard from '@/components/PaperCard';
import { fetchRandomPapers } from '@/services/paperService';
import { Paper } from '@/types/paper';
import { useQuery } from '@tanstack/react-query';

export default function Home() {
  const { data: papers = [], isLoading } = useQuery({
    queryKey: ['randomPapers'],
    queryFn: () => fetchRandomPapers(10),
  });

  return (
    <Layout>
      <div className="space-y-6">
        <h1 className="text-3xl font-bold tracking-tight">Research Papers</h1>
        
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="h-64 rounded-lg bg-muted animate-pulse" />
            ))}
          </div>
        ) : papers.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {papers.map((paper) => (
              <PaperCard key={paper.id} paper={paper} />
            ))}
          </div>
        ) : (
          <div className="text-center py-10">
            <h3 className="text-lg font-medium">No papers found</h3>
            <p className="text-muted-foreground">Try refreshing the page or check back later.</p>
          </div>
        )}
      </div>
    </Layout>
  );
}
