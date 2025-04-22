import React, { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { analyzePaper } from "@/services/paperService";
import { useRouter } from "next/navigation";

interface SearchBarPaperProps {
  paperContext: string;
}

const SearchBarPaper: React.FC<SearchBarPaperProps> = ({ paperContext }) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const router = useRouter();

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;
    setLoading(true);
    try {
      router.push(
        `/search?q=${encodeURIComponent(searchQuery)}&t=${encodeURIComponent(
          searchQuery
        )}&c=${paperContext}`
      );
    } catch (error) {
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={handleSearch}
      className="flex w-full max-w-3xl mx-auto gap-2"
    >
      <Input
        type="text"
        placeholder={`Search within \"${paperContext}\"...`}
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        className="flex-grow"
        disabled={loading}
      />
      <Button type="submit" disabled={loading}>
        Search
      </Button>
    </form>
  );
};

export default SearchBarPaper;
