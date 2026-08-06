package com.example.blog.models;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import java.time.LocalDateTime;

@Document(collection = "likes")
public class Like {
    @Id
    private String id;
    private String blogId;
    private Long authorId;
    private String authorUsername;
    private LocalDateTime createdAt;

    public Like() {}

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getBlogId() { return blogId; }
    public void setBlogId(String blogId) { this.blogId = blogId; }

    public Long getAuthorId() { return authorId; }
    public void setAuthorId(Long authorId) { this.authorId = authorId; }

    public String getAuthorUsername() { return authorUsername; }
    public void setAuthorUsername(String authorUsername) { this.authorUsername = authorUsername; }

    public LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
}
