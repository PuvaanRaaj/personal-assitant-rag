import { useState, useRef, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { queryApi } from '@/api/client';
import type { QueryResponse } from '@/types';
import { ArrowLeft, Send, Loader2, FileText, Bot, User } from 'lucide-react';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  sources?: QueryResponse['sources'];
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const queryMutation = useMutation({
    mutationFn: queryApi.ask,
    onSuccess: (data: QueryResponse) => {
      setMessages((prev) => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: data.answer,
          sources: data.sources,
        },
      ]);
    },
    onError: () => {
      setMessages((prev) => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: 'Sorry, I encountered an error processing your request. Please try again.',
        },
      ]);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || queryMutation.isPending) return;

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: 'user',
      content: input.trim(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput('');
    queryMutation.mutate(input.trim());
  };

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  return (
    <div className="h-screen flex flex-col">
      {/* Header */}
      <header className="glass border-b border-border px-6 py-4 flex items-center gap-4">
        <button
          onClick={() => navigate('/')}
          className="p-2 rounded-lg hover:bg-bg-elevated transition-colors"
        >
          <ArrowLeft className="w-5 h-5 text-text-muted" />
        </button>
        <div>
          <h1 className="text-lg font-semibold text-text">Chat with Documents</h1>
          <p className="text-text-muted text-sm">Ask questions about your uploaded documents</p>
        </div>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-3xl mx-auto space-y-6">
          {messages.length === 0 && (
            <div className="text-center py-16">
              <Bot className="w-16 h-16 text-text-muted mx-auto mb-4" />
              <h2 className="text-xl font-semibold text-text mb-2">
                Start a Conversation
              </h2>
              <p className="text-text-muted max-w-md mx-auto">
                Ask me anything about the documents you've uploaded. I'll search through
                them and provide answers with source citations.
              </p>
            </div>
          )}

          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex gap-4 ${message.role === 'user' ? 'justify-end' : ''}`}
            >
              {message.role === 'assistant' && (
                <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center shrink-0">
                  <Bot className="w-5 h-5 text-primary" />
                </div>
              )}

              <div
                className={`max-w-2xl ${
                  message.role === 'user'
                    ? 'bg-primary text-white rounded-2xl rounded-tr-md px-4 py-3'
                    : 'card'
                }`}
              >
                <p className="whitespace-pre-wrap">{message.content}</p>

                {message.sources && message.sources.length > 0 && (
                  <div className="mt-4 pt-4 border-t border-border">
                    <p className="text-text-muted text-sm mb-2">Sources:</p>
                    <div className="space-y-2">
                      {message.sources.map((source, i) => (
                        <div
                          key={i}
                          className="flex items-center gap-2 text-sm text-text-muted"
                        >
                          <FileText className="w-4 h-4 shrink-0" />
                          <span className="truncate">{source.filename}</span>
                          {source.page && (
                            <span className="text-xs bg-bg-elevated px-2 py-0.5 rounded">
                              Page {source.page}
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {message.role === 'user' && (
                <div className="w-8 h-8 rounded-lg bg-bg-elevated flex items-center justify-center shrink-0">
                  <User className="w-5 h-5 text-text-muted" />
                </div>
              )}
            </div>
          ))}

          {queryMutation.isPending && (
            <div className="flex gap-4">
              <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center shrink-0">
                <Bot className="w-5 h-5 text-primary" />
              </div>
              <div className="card">
                <div className="flex items-center gap-2 text-text-muted">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>Searching documents...</span>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>
      </div>

      {/* Input */}
      <div className="glass border-t border-border p-4">
        <form onSubmit={handleSubmit} className="max-w-3xl mx-auto flex gap-4">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask a question about your documents..."
            className="input flex-1"
            disabled={queryMutation.isPending}
          />
          <button
            type="submit"
            disabled={!input.trim() || queryMutation.isPending}
            className="btn-primary px-6 flex items-center gap-2"
          >
            {queryMutation.isPending ? (
              <Loader2 className="w-5 h-5 animate-spin" />
            ) : (
              <Send className="w-5 h-5" />
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
