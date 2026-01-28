import { useEffect, useMemo, useState } from 'react';
import type { Screen, VideoTutorial } from '../types';
import { apiRequest } from '../api/client';

interface VideoTutorialsProps {
  onNavigate: (screen: Screen) => void;
}

const difficultyLabels: Record<string, string> = {
  beginner: 'Новичок',
  intermediate: 'Средний',
  advanced: 'Продвинутый',
};

const difficultyStyles: Record<string, { bg: string; text: string }> = {
  beginner: { bg: 'bg-green-100', text: 'text-green-700' },
  intermediate: { bg: 'bg-yellow-100', text: 'text-yellow-700' },
  advanced: { bg: 'bg-red-100', text: 'text-red-700' },
};

const formatDuration = (minutes: number) => {
  if (!minutes) return '0 мин';
  const hrs = Math.floor(minutes / 60);
  const mins = minutes % 60;
  if (hrs > 0) {
    return `${hrs}:${String(mins).padStart(2, '0')} ч`;
  }
  return `${minutes} мин`;
};

export function VideoTutorials({ onNavigate }: VideoTutorialsProps) {
  const [videos, setVideos] = useState<VideoTutorial[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [query, setQuery] = useState('');
  const [selectedVideo, setSelectedVideo] = useState<VideoTutorial | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const categories = useMemo(() => {
    const items = Array.from(new Set(videos.map((video) => video.category).filter(Boolean)));
    return ['all', ...items];
  }, [videos]);

  const load = async (cat = selectedCategory, search = query) => {
    setLoading(true);
    setError(null);
    try {
      const categoryParam = cat === 'all' ? '' : cat;
      const data = await apiRequest<{ videos: any[] }>(`/videos?category=${encodeURIComponent(categoryParam)}&q=${encodeURIComponent(search)}`);
      const items = (data.videos || []).map((video) => ({
        id: video.id,
        title: video.title,
        description: video.description,
        duration: Number(video.duration || 0),
        category: video.category || 'Общее',
        difficulty: video.difficulty || 'beginner',
        url: video.url || '',
      }));
      setVideos(items);
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить видео');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const filteredVideos = useMemo(() => {
    if (selectedCategory === 'all') return videos;
    return videos.filter((video) => video.category === selectedCategory);
  }, [videos, selectedCategory]);

  const featured = filteredVideos[0] || null;

  const openVideo = (video: VideoTutorial) => {
    setSelectedVideo(video);
  };

  const handleSearch = (event: React.FormEvent) => {
    event.preventDefault();
    load(selectedCategory, query);
  };

  const playSelected = () => {
    if (selectedVideo?.url) {
      window.open(selectedVideo.url, '_blank', 'noopener');
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white p-6">
        <button
          onClick={() => onNavigate('home')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="text-2xl mb-2">Видеоуроки</h1>
        <p className="text-blue-100">Обучающие материалы от профессионалов</p>
      </div>

      <div className="p-6 space-y-6">
        <form onSubmit={handleSearch} className="flex gap-2">
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="flex-1 px-4 py-2 rounded-xl border border-gray-200"
            placeholder="Поиск по названию"
          />
          <button type="submit" className="px-4 py-2 rounded-xl bg-blue-600 text-white">
            Найти
          </button>
        </form>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        <div className="flex gap-2 overflow-x-auto pb-2 -mx-6 px-6">
          {categories.map((category) => (
            <button
              key={category}
              onClick={() => {
                setSelectedCategory(category);
                load(category, query);
              }}
              className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                selectedCategory === category
                  ? 'bg-blue-600 text-white'
                  : 'bg-white text-gray-700 border border-gray-200'
              }`}
            >
              {category === 'all' ? 'Все' : category}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Загрузка видео...</div>
        ) : featured ? (
          <div className="bg-gradient-to-br from-purple-500 to-blue-600 rounded-2xl overflow-hidden shadow-lg">
            <div className="aspect-video bg-gradient-to-br from-purple-400/30 to-blue-500/30 flex items-center justify-center">
              <div className="text-center text-white">
                <div className="bg-white/20 backdrop-blur-sm w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg className="w-10 h-10" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M8 5v14l11-7z" />
                  </svg>
                </div>
                <h3 className="text-xl mb-2">Рекомендуем</h3>
                <p className="text-sm opacity-90">{featured.category || 'Новая подборка'}</p>
              </div>
            </div>
            <div className="p-5 text-white">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <h3 className="mb-1">{featured.title}</h3>
                  <p className="text-sm text-white/80">{formatDuration(featured.duration)}</p>
                </div>
                <button
                  className="bg-white text-blue-600 px-4 py-2 rounded-lg"
                  onClick={() => openVideo(featured)}
                >
                  Смотреть
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-white rounded-2xl p-6 text-center text-gray-600">
            Видео по выбранной категории не найдены.
          </div>
        )}

        <div className="space-y-3">
          <h3>Все видео ({filteredVideos.length})</h3>
          {filteredVideos.map((tutorial) => {
            const difficulty = tutorial.difficulty || 'beginner';
            const style = difficultyStyles[difficulty] || difficultyStyles.beginner;
            return (
              <div key={tutorial.id} className="bg-white rounded-2xl overflow-hidden shadow-sm">
                <div className="flex gap-4 p-4">
                  <div className="w-32 h-20 bg-gradient-to-br from-blue-100 to-blue-50 rounded-xl flex items-center justify-center flex-shrink-0 relative">
                    <svg className="w-10 h-10 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z" />
                    </svg>
                    <div className="absolute bottom-2 right-2 bg-black/70 text-white text-xs px-2 py-0.5 rounded">
                      {formatDuration(tutorial.duration)}
                    </div>
                  </div>

                  <div className="flex-1 min-w-0">
                    <h4 className="mb-1 truncate">{tutorial.title}</h4>
                    <p className="text-sm text-gray-600 mb-2">{tutorial.description}</p>
                    <div className="flex items-center gap-2 text-xs">
                      <span className={`px-2 py-1 rounded-full ${style.bg} ${style.text}`}>
                        {difficultyLabels[difficulty] || 'Новичок'}
                      </span>
                      {tutorial.category && <span className="text-gray-500">{tutorial.category}</span>}
                    </div>
                  </div>

                  <button
                    className="text-gray-400 hover:text-gray-600"
                    onClick={() => openVideo(tutorial)}
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                    </svg>
                  </button>
                </div>
              </div>
            );
          })}
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-2xl p-5">
          <div className="flex items-start gap-3">
            <svg className="w-6 h-6 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <h4 className="text-blue-900 mb-2">Советы по просмотру</h4>
              <ul className="text-sm text-blue-700 space-y-1">
                <li>• Смотрите видео полностью перед выполнением упражнений</li>
                <li>• Повторяйте сложные моменты несколько раз</li>
                <li>• Соблюдайте правильную технику выполнения</li>
                <li>• При боли немедленно прекратите упражнение</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {selectedVideo && (
        <div className="fixed inset-0 bg-black/50 flex items-end justify-center z-50" onClick={() => setSelectedVideo(null)}>
          <div
            className="bg-white rounded-t-3xl w-full max-w-md p-6"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl">{selectedVideo.title}</h2>
              <button onClick={() => setSelectedVideo(null)} className="text-gray-400">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <p className="text-sm text-gray-600 mb-3">{selectedVideo.description}</p>
            <div className="flex items-center gap-3 text-xs text-gray-500 mb-4">
              <span>{selectedVideo.category}</span>
              <span>•</span>
              <span>{formatDuration(selectedVideo.duration)}</span>
            </div>
            <button
              onClick={playSelected}
              disabled={!selectedVideo.url}
              className="w-full bg-blue-600 text-white py-3 rounded-xl disabled:opacity-60"
            >
              {selectedVideo.url ? 'Открыть видео' : 'Ссылка недоступна'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
