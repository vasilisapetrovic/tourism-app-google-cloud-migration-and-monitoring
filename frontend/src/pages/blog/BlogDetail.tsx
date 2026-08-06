import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ReactMarkdown from 'react-markdown';
import api from '../../api';
import './Blog.scss';


interface Blog {
  id: string;
  authorId: number;
  authorUsername: string;
  title: string;
  description: string;
  createdAt: string;
  images: string[];
  likeCount?: number;
}

interface Comment {
  id: string;
  authorId: number;
  authorUsername: string;
  text: string;
  createdAt: string;
  updatedAt: string;
}

const parseDate = (dateStr: string): Date => {
  if (!dateStr) return new Date();
  const normalized = dateStr.endsWith('Z') ? dateStr : dateStr + 'Z';
  return new Date(normalized);
};

const isEdited = (createdAt: string, updatedAt: string): boolean => {
  return Math.abs(parseDate(updatedAt).getTime() - parseDate(createdAt).getTime()) > 1000;
};

export default function BlogDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [blog, setBlog] = useState<Blog | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [text, setText] = useState('');
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editText, setEditText] = useState('');
  const [hasLiked, setHasLiked] = useState(false);

  const token = localStorage.getItem('token');
  const currentUserId = token ? JSON.parse(atob(token.split('.')[1])).id : null;

  useEffect(() => {
    api.get(`/blogs/${id}`).then((res) => setBlog(res.data));
    api.get(`/comments/${id}`).then((res) => setComments(res.data));

    api.get(`/likes/check/${id}`).then((res) => {
      setHasLiked(res.data.hasLiked);
    }).catch(() => setHasLiked(false));
  }, [id]);

  const handleComment = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      const res = await api.post(`/comments/${id}`, { text });
      setComments((prev) => [res.data, ...prev]);
      setText('');
    } catch {
      setError('Failed to post comment');
    }
  };

  const handleEdit = async (commentId: string) => {
    try {
      const res = await api.put(`/comments/${commentId}`, { text: editText });
      setComments((prev) =>
        prev.map((c) =>
          c.id === commentId
            ? { ...c, text: editText, updatedAt: res.data?.updatedAt ?? new Date().toISOString() }
            : c
        )
      );
      setEditingId(null);
    } catch {
      setError('Failed to edit comment');
    }
  };

  const handleLike = async () => {
    if (!blog) return;
    try {
      if (hasLiked) {
        const response = await api.delete(`/likes/${id}`);
        setBlog({ ...blog, likeCount: response.data.likeCount });
        setHasLiked(false);
      } else {
        const response = await api.post(`/likes/${id}`, {});
        setBlog({ ...blog, likeCount: response.data.likeCount });
        setHasLiked(true);
      }
    } catch (error: any) {
      if (error.response?.status === 401) {
        navigate('/login');
      } else {
        console.error('Like error:', error);
      }
    }
  };

  if (!blog) return <p>Loading...</p>;

  return (
    <>
      <div className="blog-hero">
        <h1>{blog.title}</h1>
      </div>
      <div className="blog-page">
        <button className="back-btn" onClick={() => navigate('/blogs')}>← Back</button>

        <div className="blog-card">
          <div className="blog-card-wrapper">
            <div className="blog-card-content">
              <p className="blog-meta">{blog.authorUsername} · {parseDate(blog.createdAt).toLocaleString()}</p>
              <div className="blog-content">
                <ReactMarkdown>{blog.description}</ReactMarkdown>
              </div>
              {blog.images.length > 0 && (
                <div className="blog-images">
                  {blog.images.map((url, i) => (
                    <img key={i} src={url} alt={`blog-img-${i}`} />
                  ))}
                </div>
              )}
            </div>
            <div className="blog-card-sidebar">
              <button
                className={`like-btn ${hasLiked ? 'liked' : ''}`}
                onClick={handleLike}
              >
                <span className="like-icon">♥</span>
                <span className="like-count">{blog.likeCount || 0}</span>
              </button>
            </div>
          </div>
        </div>

        <div className="comments-section">
          <h3>Comments ({comments.length})</h3>

          <form className="comment-form" onSubmit={handleComment}>
            {error && <p className="auth-error">{error}</p>}
            <textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Write a comment..."
              rows={3}
              required
            />
            <button className="btn" type="submit">Post Comment</button>
          </form>

          {comments.length === 0 ? (
            <p className="no-comments">No comments yet. Be the first!</p>
          ) : (
            comments.map((c) => (
              <div key={c.id} className="comment-card">

                <div className="comment-card-header">
                  <div className="comment-card-left">
                    <div className="comment-avatar">
                      {c.authorUsername.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <p className="comment-author">{c.authorUsername}</p>
                      <p className="comment-meta">
                        {parseDate(c.updatedAt).toLocaleString()}
                        {isEdited(c.createdAt, c.updatedAt) && (
                          <span className="edited-badge">edited</span>
                        )}
                      </p>
                    </div>
                  </div>

                  {c.authorId === currentUserId && editingId !== c.id && (
                    <button
                      className="btn-edit"
                      onClick={() => {
                        setEditingId(c.id);
                        setEditText(c.text);
                      }}
                    >
                      Edit
                    </button>
                  )}
                </div>

                {editingId === c.id ? (
                  <div className="comment-edit-area">
                    <textarea
                      value={editText}
                      onChange={(e) => setEditText(e.target.value)}
                      rows={2}
                    />
                    <div className="comment-edit-actions">
                      <button className="btn" onClick={() => handleEdit(c.id)}>Save</button>
                      <button className="btn-cancel" onClick={() => setEditingId(null)}>Cancel</button>
                    </div>
                  </div>
                ) : (
                  <p className="comment-text">{c.text}</p>
                )}

              </div>
            ))
          )}
        </div>
      </div>
    </>
  );
}