package com.example.blog.repositories;

import com.example.blog.models.Comment;
import org.springframework.data.mongodb.repository.MongoRepository;
import java.util.List;

public interface CommentRepository extends MongoRepository<Comment, String> {
    List<Comment> findByBlogId(String blogId);
    List<Comment> findByBlogIdOrderByCreatedAtDesc(String blogId);
}