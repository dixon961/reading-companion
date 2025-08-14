import React from 'react';

interface MarkdownRendererProps {
  markdown: string;
}

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ markdown }) => {
  // Simple markdown to HTML converter for our specific use case
  const renderMarkdown = (md: string): string => {
    let html = md;
    
    // Convert headers
    html = html.replace(/^# (.*$)/gm, '<h1>$1</h1>');
    html = html.replace(/^## (.*$)/gm, '<h2>$1</h2>');
    html = html.replace(/^### (.*$)/gm, '<h3>$1</h3>');
    
    // Convert blockquotes
    html = html.replace(/^> (.*)$/gm, '<blockquote class="markdown-blockquote">$1</blockquote>');
    
    // Convert bold text
    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    
    // Convert italics
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');
    
    // Convert paragraphs (lines that aren't already wrapped in block-level tags)
    html = html.replace(/^(?!<(h1|h2|h3|blockquote|ul|ol|li|pre|code))(.+)$/gm, '<p>$1</p>');
    
    // Handle line breaks within paragraphs
    html = html.replace(/\n/g, '<br>');
    
    return html;
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