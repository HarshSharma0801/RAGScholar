import React, { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';
import { analyzePaper } from '@/services/paperService';

interface SearchBarProps {
  paperContext?: string;
}

const SearchBar: React.FC<SearchBarProps> = ({ paperContext }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const router = useRouter();

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!searchQuery.trim()) return;
    
    try {
      await analyzePaper({
        selectedText: searchQuery,
        paperContext: paperContext || '',
        searchQuery: searchQuery,
        customPrompt: '',
      });
      
      // If on the home page, redirect to search results
      if (!paperContext) {
        router.push(`/search?q=${encodeURIComponent(searchQuery)}`);
      }
      // If on a paper page, results will be handled by the parent component
    } catch (error) {
      console.error('Search error:', error);
    }
  };

  return (
    <form onSubmit={handleSearch} className="flex w-full max-w-3xl mx-auto gap-2">
      <Input
        type="text"
        placeholder="Search papers..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        className="flex-grow"
      />
      <Button type="submit">Search</Button>
    </form>
  );
};

export default SearchBar;
