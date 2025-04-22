import axios from "axios";
import { Paper } from "@/types/paper";

const API_URL = "http://localhost:8040";


interface Author {
  Name: string;
}

export interface PaperById {
  id: string;
  title: string;
  summary: string;
  authors: Author[];
  published: string;
  updated: string;
  comment?: string;
  links: {
    Href: string;
    Rel: string;
    Type: string;
  }[];
  categories?: string[];
  journalRef?: string;
}


export const fetchRandomPapers = async (
  count: number = 10
): Promise<Paper[]> => {
  try {
    const response = await axios.get(`${API_URL}/`);
    return response.data.papers;
  } catch (error) {
    console.error("Error fetching random papers:", error);
    return [];
  }
};

export const fetchPaperById = async (id: string): Promise<PaperById | null> => {
  try {
    const response = await axios.get(`${API_URL}/paper/${id}`);
    if (response.data) {
      return response.data.paper;
    }
    return null;
  } catch (error) {
    console.error("Error fetching random papers:", error);
    return null;
  }
};

export const analyzePaper = async (params: {
  selectedText?: string;
  paperContext?: string;
  searchQuery?: string;
  customPrompt?: string;
}): Promise<any> => {
  try {
    const response = await axios.post(`${API_URL}/analyze`, params);
    return response.data;
  } catch (error) {
    console.error("Error analyzing paper:", error);
    throw error;
  }
};
