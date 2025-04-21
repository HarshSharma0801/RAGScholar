'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Layout from '@/components/Layout';
import { Paper } from '@/types/paper';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { analyzePaper } from '@/services/paperService';

// Mock data for a single paper (since the API is not available)
const mockPaperData: Record<string, Paper> = {
  '1': {
    id: '1',
    title: 'Review on Efficient Strategies for Coordinated Motion and Tracking in Swarm Robotics',
    abstract: 'This paper presents a comprehensive review of strategies for coordinated motion and tracking in swarm robotics. We analyze various approaches including distributed algorithms, consensus-based methods, and bio-inspired techniques. The review highlights the trade-offs between computational complexity, communication overhead, and robustness in different coordination strategies.',
    authors: ['Jane Smith', 'John Doe', 'Alice Johnson'],
    year: '2023',
    doi: '10.1234/5678.9012',
    citations: 42,
    keywords: ['swarm robotics', 'coordination', 'distributed algorithms', 'tracking']
  },
  '2': {
    id: '2',
    title: 'Machine Learning Approaches for Predictive Maintenance in Industrial IoT',
    abstract: 'This study explores machine learning techniques for predictive maintenance in Industrial Internet of Things (IIoT) environments. We compare various algorithms including random forests, support vector machines, and deep learning approaches for failure prediction and maintenance scheduling.',
    authors: ['Robert Brown', 'Sarah Lee'],
    year: '2022',
    doi: '10.5678/1234.5678',
    citations: 28,
    keywords: ['predictive maintenance', 'machine learning', 'industrial IoT', 'failure prediction']
  },
  // Add more mock papers as needed
};

export default function PaperPage() {
  const params = useParams();
  const paperId = params.id as string;
  
  const [paper, setPaper] = useState<Paper | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchResults, setSearchResults] = useState<any | null>(null);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    // In a real application, this would fetch the paper from an API
    // For now, we'll use mock data
    const fetchPaper = async () => {
      try {
        // Simulate API call delay
        await new Promise(resolve => setTimeout(resolve, 500));
        
        const paperData = mockPaperData[paperId];
        if (paperData) {
          setPaper(paperData);
        }
      } catch (error) {
        console.error('Failed to fetch paper:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchPaper();
  }, [paperId]);

  const handleSearch = async (query: string) => {
    if (!paper) return;
    
    try {
      setSearching(true);
      const results = await analyzePaper({
        selectedText: query,
        paperContext: paper.title,
        searchQuery: '',
        customPrompt: '',
      });
      
      setSearchResults(results);
    } catch (error) {
      console.error('Error analyzing paper:', error);
    } finally {
      setSearching(false);
    }
  };

  if (loading) {
    return (
      <Layout>
        <div className="space-y-4 animate-pulse">
          <div className="h-8 bg-muted rounded w-3/4"></div>
          <div className="h-4 bg-muted rounded w-1/2"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </Layout>
    );
  }

  if (!paper) {
    return (
      <Layout>
        <div className="text-center py-10">
          <h2 className="text-2xl font-bold">Paper Not Found</h2>
          <p className="text-muted-foreground mt-2">The requested paper could not be found.</p>
          <Button className="mt-4" onClick={() => window.history.back()}>Go Back</Button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout paperContext={paper.title}>
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{paper.title}</CardTitle>
            <CardDescription>
              {paper.authors?.join(', ')} • {paper.year}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold mb-2">Abstract</h3>
              <p className="text-muted-foreground">{paper.abstract}</p>
            </div>
            
            {paper.keywords && paper.keywords.length > 0 && (
              <div>
                <h3 className="text-lg font-semibold mb-2">Keywords</h3>
                <div className="flex flex-wrap gap-2">
                  {paper.keywords.map((keyword, index) => (
                    <span 
                      key={index} 
                      className="px-2 py-1 bg-secondary text-secondary-foreground rounded-md text-sm"
                    >
                      {keyword}
                    </span>
                  ))}
                </div>
              </div>
            )}
            
            {paper.doi && (
              <div>
                <h3 className="text-lg font-semibold mb-2">DOI</h3>
                <p className="text-muted-foreground">{paper.doi}</p>
              </div>
            )}
            
            {paper.citations !== undefined && (
              <div>
                <h3 className="text-lg font-semibold mb-2">Citations</h3>
                <p className="text-muted-foreground">{paper.citations}</p>
              </div>
            )}
          </CardContent>
        </Card>

        {searching && (
          <div className="text-center py-4">
            <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]"></div>
            <p className="mt-2">Analyzing paper...</p>
          </div>
        )}

        {searchResults && (
          <Card>
            <CardHeader>
              <CardTitle>Analysis Results</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="bg-muted p-4 rounded-md overflow-auto">
                {JSON.stringify(searchResults, null, 2)}
              </pre>
            </CardContent>
          </Card>
        )}
      </div>
    </Layout>
  );
}
