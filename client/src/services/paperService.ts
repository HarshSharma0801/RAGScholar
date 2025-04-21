import axios from 'axios';
import { Paper } from '@/types/paper';

const API_URL = 'http://localhost:8040';

export const fetchRandomPapers = async (count: number = 10): Promise<Paper[]> => {
  try {
    const response = await axios.get(`${API_URL}/`);
    return response.data.papers
  } catch (error) {
    console.error('Error fetching random papers:', error);
    return [];
  }
};

export const searchPapers = async (query: string): Promise<Paper[]> => {
  try {
    const response = await axios.post(`${API_URL}/analyze`, {
      searchQuery: query,
      selectedText: '',
      paperContext: '',
      customPrompt: '',
    });
    return response.data;
  } catch (error) {
    console.error('Error searching papers:', error);
    return [];
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
    console.error('Error analyzing paper:', error);
    throw error;
  }
};
