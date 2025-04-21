'use client';

import { useSearchParams } from 'next/navigation';
import Layout from '@/components/Layout';
import PaperCard from '@/components/PaperCard';
import { Paper } from '@/types/paper';
import { searchPapers } from '@/services/paperService';
import { useQuery } from '@tanstack/react-query';

export default function SearchPage() {
  const searchParams = useSearchParams();
  const query = searchParams.get('q') || '';
  
  const { data: papers = [], isLoading } = useQuery({
    queryKey: ['searchPapers', query],
    queryFn: () => searchPapers(query),
    enabled: !!query,
  });

  return (
    <Layout>
      <div className="space-y-6">
        <h1 className="text-3xl font-bold tracking-tight">
          Search Results: <span className="text-muted-foreground">{query}</span>
        </h1>
        
        {isLoading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 5 }).map((_, i) => (
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
            <h3 className="text-lg font-medium">No results found</h3>
            <p className="text-muted-foreground">
              Try searching with different keywords or check your spelling.
            </p>
          </div>
        )}
      </div>
    </Layout>
  );
}
