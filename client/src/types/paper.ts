export interface Paper {
  id: string;
  title: string;
  abstract?: string;
  authors?: { Name: string }[];
  year?: string;
  url?: string;
  doi?: string;
  citations?: number;
  content?: string;
  keywords?: string[];
  categories?: string[];
  published?: string;
}
