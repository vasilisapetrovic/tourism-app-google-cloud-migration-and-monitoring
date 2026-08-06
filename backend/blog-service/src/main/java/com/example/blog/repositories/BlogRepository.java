package com.example.blog.repositories;

import com.example.blog.models.Blog;
import org.springframework.data.mongodb.repository.MongoRepository;
import java.util.List;

public interface BlogRepository extends MongoRepository<Blog, String> {
    List<Blog> findByAuthorIdIn(List<Long> authorIds);
    List<Blog> findByAuthorId(Long authorId);
}