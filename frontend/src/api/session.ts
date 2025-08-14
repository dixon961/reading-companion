// api/session.ts
export interface CreateSessionResponse {
  session_id: string;
  name: string;
  total_highlights: number;
  next_step: {
    highlight_index: number;
    highlight_text: string;
    question: string;
  };
}

export interface ProcessAnswerResponse {
  status?: string;
  message?: string;
  next_step?: {
    highlight_index: number;
    highlight_text: string;
    question: string;
  };
}

export interface ProcessAnswerRequest {
  highlight_index: number;
  user_answer: string;
}

export interface SessionMetadata {
  id: string;
  name: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface SessionData {
  id: string;
  name: string;
  status: string;
  total_highlights: number;
  next_step: {
    highlight_index: number;
    highlight_text: string;
    question: string;
  };
}

export interface RegenerateQuestionRequest {
  highlight_index: number;
}

export interface RegenerateQuestionResponse {
  new_question: string;
}

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090/api';

export const listSessions = async (): Promise<SessionMetadata[]> => {
  const response = await fetch(`${API_BASE_URL}/sessions`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to list sessions: ${errorText}`);
  }

  return response.json();
};

export const createSession = async (file: File, sessionName?: string): Promise<CreateSessionResponse> => {
  const formData = new FormData();
  formData.append('file', file);
  if (sessionName) {
    formData.append('session_name', sessionName);
  }

  const response = await fetch(`${API_BASE_URL}/sessions`, {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const errorText = await response.text();
    // Handle specific error cases
    if (response.status === 400) {
      if (errorText.includes("Файл пуст")) {
        throw new Error("Файл пуст или имеет неверный формат");
      } else if (errorText.includes("Для начала сессии необходимо как минимум 3 пометки")) {
        throw new Error("Для начала сессии необходимо как минимум 3 пометки");
      }
    }
    throw new Error(`Failed to create session: ${errorText}`);
  }

  return response.json();
};

export interface UpdateSessionNameRequest {
  name: string;
}

export interface UpdateSessionNameResponse {
  id: string;
  name: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export const updateSessionName = async (sessionId: string, data: UpdateSessionNameRequest): Promise<UpdateSessionNameResponse> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to update session name: ${errorText}`);
  }

  return response.json();
};

export const deleteSession = async (sessionId: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to delete session: ${errorText}`);
  }
};

export const getSession = async (sessionId: string): Promise<SessionData> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to get session: ${errorText}`);
  }

  return response.json();
};

export const processAnswer = async (sessionId: string, data: ProcessAnswerRequest): Promise<ProcessAnswerResponse> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}/process`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorText = await response.text();
    // Handle 503 Service Unavailable errors
    if (response.status === 503) {
      throw new Error("LLM service unavailable. Please try again later.");
    }
    throw new Error(`Failed to process answer: ${errorText}`);
  }

  return response.json();
};

export const regenerateQuestion = async (sessionId: string, data: RegenerateQuestionRequest): Promise<RegenerateQuestionResponse> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}/regenerate_question`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const errorText = await response.text();
    // Handle 503 Service Unavailable errors
    if (response.status === 503) {
      throw new Error("LLM service unavailable. Please try again later.");
    }
    throw new Error(`Failed to regenerate question: ${errorText}`);
  }

  return response.json();
};

export const exportSession = async (sessionId: string): Promise<string> => {
  const response = await fetch(`${API_BASE_URL}/sessions/${sessionId}/export`, {
    method: 'GET',
    headers: {
      'Accept': 'text/markdown',
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to export session: ${errorText}`);
  }

  return response.text();
};

// Utility function to download content as a file
export const downloadFile = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/markdown;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  
  document.body.appendChild(link);
  link.click();
  
  // Clean up
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};