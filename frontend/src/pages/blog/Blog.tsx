import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
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

export default function Blogs() {
  const [blogs, setBlogs] = useState<Blog[]>([]);
  const [userLikes, setUserLikes] = useState<Set<string>>(new Set());
  const navigate = useNavigate();

  useEffect(() => {
    const fetchBlogs = async () => {
      try {
        // Fetch both my blogs and followed users' blogs
        const [myBlogsRes, followedBlogsRes] = await Promise.all([
          api.get('/blogs/my').catch(() => ({ data: [] })),
          api.get('/blogs/followed').catch(() => ({ data: [] }))
        ]);

        // Combine and remove duplicates
        const allBlogs = [...myBlogsRes.data, ...followedBlogsRes.data];
        const uniqueBlogs = Array.from(
          new Map(allBlogs.map(blog => [blog.id, blog])).values()
        );
        // Sort by creation date (newest first)
        uniqueBlogs.sort((a, b) => 
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );

        setBlogs(uniqueBlogs);
        
        const likes = new Set<string>();
        for (const blog of uniqueBlogs) {
          try {
            const likeRes = await api.get(`/likes/check/${blog.id}`);
            if (likeRes.data.hasLiked) {
              likes.add(blog.id);
            }
          } catch {
          }
        }
        setUserLikes(likes);
      } catch (error) {
        console.error('Failed to fetch blogs:', error);
      }
    };
    
    fetchBlogs();
  }, []);

  const handleLikeClick = async (e: React.MouseEvent, blogId: string) => {
    e.stopPropagation();
    
    try {
      if (userLikes.has(blogId)) {
        const response = await api.delete(`/likes/${blogId}`);
        setBlogs((prev) =>
          prev.map((b) =>
            b.id === blogId
              ? { ...b, likeCount: response.data.likeCount }
              : b
          )
        );
        setUserLikes((prev) => {
          const newSet = new Set(prev);
          newSet.delete(blogId);
          return newSet;
        });
      } else {
        const response = await api.post(`/likes/${blogId}`, {});
        setBlogs((prev) =>
          prev.map((b) =>
            b.id === blogId
              ? { ...b, likeCount: response.data.likeCount }
              : b
          )
        );
        setUserLikes((prev) => new Set(prev).add(blogId));
      }
    } catch (error: any) {
      // Ako je 401 (Unauthorized), preusmeri na login
      if (error.response?.status === 401) {
        navigate('/login');
      } else {
        console.error('Like error:', error);
      }
    }
  };

  return (
    <>
      <div className="blog-hero">
        <h1>Blogs</h1>
      </div>
      <div className="blog-page">
        <div className="blog-header">
          <Link to="/blogs/create" className="btn">New Blog</Link>
        </div>
        {blogs.length === 0 ? (
          <p>No blogs yet.</p>
        ) : (
          blogs.map((blog) => (
            <div key={blog.id} className="blog-card" onClick={() => navigate(`/blogs/${blog.id}`)} style={{ cursor: 'pointer' }}>
              <div className="blog-card-wrapper">
                <div className="blog-card-content">
                  <h3>{blog.title}</h3>
                  <p className="blog-meta">{blog.authorUsername} · {new Date(blog.createdAt).toLocaleDateString()}</p>
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
                    className={`like-btn ${userLikes.has(blog.id) ? 'liked' : ''}`}
                    onClick={(e) => handleLikeClick(e, blog.id)}
                  >
                    <span className="like-icon">♥</span>
                    <span className="like-count">{blog.likeCount || 0}</span>
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </>
  );
}