import React from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { Paper } from '@/types/paper';

interface PaperCardProps {
  paper: Paper;
}

const PaperCard: React.FC<PaperCardProps> = ({ paper }) => {
  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle className="line-clamp-2">{paper.title}</CardTitle>
        <CardDescription>
          {paper.authors?.join(', ') || 'Unknown Authors'} • {paper.year || 'N/A'}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-grow">
        <p className="text-sm text-muted-foreground line-clamp-4">
          {paper.abstract || 'No abstract available'}
        </p>
      </CardContent>
      <CardFooter>
        <Link href={`/paper/${paper.title}`} className="w-full">
          <Button variant="outline" className="w-full">View Paper</Button>
        </Link>
      </CardFooter>
    </Card>
  );
};

export default PaperCard;
