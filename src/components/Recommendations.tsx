import { useEffect, useState } from 'react';
import type { Recommendation, Screen } from '../types';
import { apiRequest } from '../api/client';

interface RecommendationsProps {
  onNavigate: (screen: Screen) => void;
}

export function Recommendations({ onNavigate }: RecommendationsProps) {
  const [category, setCategory] = useState('');
  const [categories, setCategories] = useState<string[]>(['Все']);
  const [tips, setTips] = useState<any[]>([]);
  const [articles, setArticles] = useState<Recommendation[]>([]);
  const [videos, setVideos] = useState<any[]>([]);
  const [selected, setSelected] = useState<Recommendation | null>(null);

  const load = async (cat = category) => {
    const data = await apiRequest<any>(`/recommendations?category=${encodeURIComponent(cat)}`);
    setCategories(data.categories || ['Все']);
    setTips(data.tips || []);
    const items = (data.articles || []).map((article: any) => ({
      id: article.id,
      title: article.title,
      category: article.category,
      readTime: article.read_time,
      icon: article.icon,
      excerpt: article.excerpt,
      body: article.body,
      bookmarked: article.bookmarked,
    }));
    setArticles(items);
    setVideos(data.videos || []);
  };

  useEffect(() => {
    load();
  }, []);

  const openArticle = async (id: string) => {
    const data = await apiRequest<any>(`/recommendations/${id}`);
    const article = data.article;
    setSelected({
      id: article.id,
      title: article.title,
      category: article.category,
      readTime: article.read_time,
      icon: article.icon,
      excerpt: article.excerpt,
      body: article.body,
      bookmarked: article.bookmarked,
    });
  };

  const toggleBookmark = async () => {
    if (!selected) return;
    if (selected.bookmarked) {
      await apiRequest(`/recommendations/${selected.id}/bookmark`, { method: 'DELETE' });
      setSelected({ ...selected, bookmarked: false });
    } else {
      await apiRequest(`/recommendations/${selected.id}/bookmark`, { method: 'POST' });
      setSelected({ ...selected, bookmarked: true });
    }
    await load(category);
  };

  const startPractice = async () => {
    if (!selected) return;
    await apiRequest(`/recommendations/${selected.id}/practice`, { method: 'POST' });
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-gradient-to-br from-green-600 to-teal-600 text-white p-6 rounded-b-3xl">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('home')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Рекомендации</h1>
        </div>
        <p className="text-green-100 text-sm">Советы и материалы для эффективных тренировок</p>
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Быстрые советы</h3>
          <div className="grid grid-cols-2 gap-3">
            {tips.map((tip) => (
              <div key={tip.title} className="bg-gradient-to-br from-blue-50 to-purple-50 rounded-xl p-4">
                <div className="text-3xl mb-2">{tip.icon}</div>
                <h4 className="text-sm mb-1">{tip.title}</h4>
                <p className="text-xs text-gray-600">{tip.description}</p>
              </div>
            ))}
          </div>
        </div>

        <div>
          <h3 className="mb-3">Статьи и материалы</h3>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {categories.map((item) => (
              <button
                key={item}
                onClick={() => {
                  setCategory(item === 'Все' ? '' : item);
                  load(item === 'Все' ? '' : item);
                }}
                className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                  (category === '' && item === 'Все') || category === item
                    ? 'bg-green-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-200'
                }`}
              >
                {item}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          {articles.map((article) => (
            <button
              key={article.id}
              onClick={() => openArticle(article.id)}
              className="bg-white rounded-2xl p-4 shadow-sm cursor-pointer hover:shadow-md transition-shadow block text-left"
            >
              <div className="flex gap-4">
                <div className="bg-gradient-to-br from-green-100 to-teal-100 rounded-xl w-20 h-20 flex items-center justify-center text-3xl flex-shrink-0">
                  {article.icon}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs px-2 py-1 rounded-full bg-green-50 text-green-700">{article.category}</span>
                    <span className="text-xs text-gray-500">{article.readTime} мин чтения</span>
                  </div>
                  <h4 className="mb-1">{article.title}</h4>
                  <p className="text-sm text-gray-600">{article.excerpt}</p>
                </div>
                <svg className="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </button>
          ))}
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Видео уроки</h3>
          <div className="space-y-3">
            {videos.map((video: any) => (
              <div key={video.id} className="flex items-center gap-3 p-3 bg-gray-50 rounded-xl">
                <div className="bg-red-100 rounded-lg w-12 h-12 flex items-center justify-center flex-shrink-0">
                  <svg className="w-6 h-6 text-red-600" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                </div>
                <div className="flex-1">
                  <div className="text-sm mb-1">{video.title}</div>
                  <div className="text-xs text-gray-600">{video.duration} мин</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {selected && (
        <div className="fixed inset-0 bg-black/50 flex items-end justify-center z-50">
          <div className="bg-white rounded-t-3xl w-full max-w-md overflow-y-auto" style={{ maxHeight: '85vh' }}>
            <div className="p-6 relative">
              <button
                onClick={() => setSelected(null)}
                className="absolute top-4 right-4 bg-gray-100 rounded-full p-2"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
              <div className="bg-gradient-to-br from-green-100 to-teal-100 rounded-2xl w-full h-32 flex items-center justify-center text-5xl mb-4">
                {selected.icon}
              </div>
              <div className="flex items-center gap-2 mb-3">
                <span className="text-xs px-3 py-1 rounded-full bg-green-50 text-green-700">{selected.category}</span>
                <span className="text-xs text-gray-500 flex items-center gap-1">
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {selected.readTime} мин
                </span>
              </div>
              <h2 className="mb-3">{selected.title}</h2>
              <p className="text-gray-600 mb-6">{selected.excerpt}</p>
              <div className="text-sm text-gray-600 space-y-3">
                <p>{selected.body}</p>
              </div>
              <div className="flex gap-3 mt-6">
                <button className="flex-1 bg-green-600 text-white py-3 rounded-xl" onClick={startPractice}>
                  Начать практику
                </button>
                <button className="px-4 py-3 bg-gray-100 rounded-xl" onClick={toggleBookmark}>
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
