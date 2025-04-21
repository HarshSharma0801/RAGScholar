export interface Paper {
  id: string;
  title: string;
  abstract?: string;
  authors?: string[];
  year?: string;
  url?: string;
  doi?: string;
  citations?: number;
  content?: string;
  keywords?: string[];
}
