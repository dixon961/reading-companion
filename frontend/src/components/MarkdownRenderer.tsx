import React from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';

interface MarkdownRendererProps {
  markdown: string;
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ markdown }) => {
  // Configure marked options for better rendering
  marked.setOptions({
    breaks: true,
    gfm: true,
  });

  // Convert markdown to HTML safely
  const renderMarkdown = (md: string): string => {
    try {
      // Parse markdown to HTML
      const rawHtml = marked.parse(md) as string;
      
      // Sanitize HTML to prevent XSS attacks
      const sanitizedHtml = DOMPurify.sanitize(rawHtml);
      
      return sanitizedHtml;
    } catch (error) {
      console.error('Error parsing markdown:', error);
      // Fallback to simple text rendering
      return `<pre>${md.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>`;
    }
  };

  // Render the markdown content
  const renderedHtml = renderMarkdown(markdown);

  return (
    <div 
      className="markdown-renderer"
      dangerouslySetInnerHTML={{ __html: renderedHtml }}
    />
  );
};

export default MarkdownRenderer;