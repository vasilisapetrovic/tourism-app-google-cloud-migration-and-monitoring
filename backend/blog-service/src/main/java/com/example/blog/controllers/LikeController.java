package com.example.blog.controllers;

import com.example.blog.models.Like;
import com.example.blog.models.Blog;
import com.example.blog.repositories.LikeRepository;
import com.example.blog.repositories.BlogRepository;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.LocalDateTime;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/likes")
@CrossOrigin(origins = "http://localhost:3000")
public class LikeController {

    private final LikeRepository likeRepository;
    private final BlogRepository blogRepository;

    public LikeController(LikeRepository likeRepository, BlogRepository blogRepository) {
        this.likeRepository = likeRepository;
        this.blogRepository = blogRepository;
    }

    @PostMapping("/{blogId}")
    public ResponseEntity<Map<String, Object>> addLike(@PathVariable String blogId,
                                                       @RequestHeader(value = "Authorization", required = false) String authHeader) {
        // Provera da li je korisnik autentifikovan
        if (authHeader == null || authHeader.trim().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "User must be logged in to like a blog"));
        }

        try {
            Long authorId = extractUserIdFromToken(authHeader);
            String username = extractUsernameFromToken(authHeader);

            Optional<Like> existingLike = likeRepository.findByBlogIdAndAuthorId(blogId, authorId);
            if (existingLike.isPresent()) {
                return ResponseEntity.badRequest().body(Map.of("error", "Already liked"));
            }

            Optional<Blog> blog = blogRepository.findById(blogId);
            if (blog.isEmpty()) {
                return ResponseEntity.notFound().build();
            }

            Like like = new Like();
            like.setBlogId(blogId);
            like.setAuthorId(authorId);
            like.setAuthorUsername(username);
            like.setCreatedAt(LocalDateTime.now());

            likeRepository.save(like);

            Blog blogObj = blog.get();
            blogObj.setLikeCount((int) likeRepository.countByBlogId(blogId));
            blogRepository.save(blogObj);

            return ResponseEntity.status(HttpStatus.CREATED).body(Map.of("likeCount", blogObj.getLikeCount()));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "Invalid or expired token"));
        }
    }

    @DeleteMapping("/{blogId}")
    public ResponseEntity<Map<String, Object>> removeLike(@PathVariable String blogId,
                                                          @RequestHeader(value = "Authorization", required = false) String authHeader) {
        // Provera da li je korisnik autentifikovan
        if (authHeader == null || authHeader.trim().isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "User must be logged in to unlike a blog"));
        }

        try {
            Long authorId = extractUserIdFromToken(authHeader);

            Optional<Like> like = likeRepository.findByBlogIdAndAuthorId(blogId, authorId);
            if (like.isEmpty()) {
                return ResponseEntity.notFound().build();
            }

            likeRepository.deleteByBlogIdAndAuthorId(blogId, authorId);

            Optional<Blog> blog = blogRepository.findById(blogId);
            if (blog.isPresent()) {
                Blog blogObj = blog.get();
                blogObj.setLikeCount((int) likeRepository.countByBlogId(blogId));
                blogRepository.save(blogObj);
                return ResponseEntity.ok(Map.of("likeCount", blogObj.getLikeCount()));
            }

            return ResponseEntity.notFound().build();
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("error", "Invalid or expired token"));
        }
    }

    @GetMapping("/check/{blogId}")
    public ResponseEntity<Map<String, Object>> checkUserLike(@PathVariable String blogId,
                                                             @RequestHeader(value = "Authorization", required = false) String authHeader) {
        long likeCount = likeRepository.countByBlogId(blogId);
        
        // Ako nema autentifikacije, samo vrati broj lajkova
        if (authHeader == null || authHeader.trim().isEmpty()) {
            return ResponseEntity.ok(Map.of(
                "hasLiked", false,
                "likeCount", likeCount
            ));
        }

        try {
            Long authorId = extractUserIdFromToken(authHeader);
            Optional<Like> like = likeRepository.findByBlogIdAndAuthorId(blogId, authorId);
            boolean hasLiked = like.isPresent();

            return ResponseEntity.ok(Map.of(
                "hasLiked", hasLiked,
                "likeCount", likeCount
            ));
        } catch (Exception e) {
            return ResponseEntity.ok(Map.of(
                "hasLiked", false,
                "likeCount", likeCount
            ));
        }
    }

    private Long extractUserIdFromToken(String authHeader) {
        String token = authHeader.replace("Bearer ", "");
        String[] parts = token.split("\\.");
        String payload = new String(java.util.Base64.getUrlDecoder().decode(parts[1]));
        String idStr = payload.split("\"id\":")[1].split("[,}]")[0];
        return Long.parseLong(idStr.trim());
    }

    private String extractUsernameFromToken(String authHeader) {
        String token = authHeader.replace("Bearer ", "");
        String[] parts = token.split("\\.");
        String payload = new String(java.util.Base64.getUrlDecoder().decode(parts[1]));
        return payload.split("\"username\":\"")[1].split("\"")[0];
    }
}
