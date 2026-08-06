package com.example.blog.repositories;

import com.example.blog.models.Like;
import org.springframework.data.mongodb.repository.MongoRepository;
import java.util.List;
import java.util.Optional;

public interface LikeRepository extends MongoRepository<Like, String> {
    List<Like> findByBlogId(String blogId);
    Optional<Like> findByBlogIdAndAuthorId(String blogId, Long authorId);
    long countByBlogId(String blogId);
    void deleteByBlogIdAndAuthorId(String blogId, Long authorId);
}
