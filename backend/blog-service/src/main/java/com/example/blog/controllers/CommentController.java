package com.example.blog.controllers;

import com.example.blog.models.Comment;
import com.example.blog.repositories.CommentRepository;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/comments")
@CrossOrigin(origins = "http://localhost:3000")
public class CommentController {

    private final CommentRepository commentRepository;

    public CommentController(CommentRepository commentRepository) {
        this.commentRepository = commentRepository;
    }

    @PostMapping("/{blogId}")
    public ResponseEntity<Comment> create(@PathVariable String blogId,
                                          @RequestBody java.util.Map<String, String> body,
                                          @RequestHeader("Authorization") String authHeader) {
        Long authorId = extractUserIdFromToken(authHeader);
        String username = extractUsernameFromToken(authHeader);

        Comment comment = new Comment();
        comment.setBlogId(blogId);
        comment.setAuthorId(authorId);
        comment.setAuthorUsername(username);
        comment.setText(body.get("text"));
        comment.setCreatedAt(LocalDateTime.now());
        comment.setUpdatedAt(LocalDateTime.now());

        Comment saved = commentRepository.save(comment);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @GetMapping("/{blogId}")
    public ResponseEntity<List<Comment>> getByBlog(@PathVariable String blogId) {
        return ResponseEntity.ok(commentRepository.findByBlogIdOrderByCreatedAtDesc(blogId));
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

    @PutMapping("/{commentId}")
    public ResponseEntity<Comment> update(@PathVariable String commentId,
                                          @RequestBody java.util.Map<String, String> body,
                                          @RequestHeader("Authorization") String authHeader) {
        Long userId = extractUserIdFromToken(authHeader);

        return commentRepository.findById(commentId)
                .map(comment -> {
                    if (!comment.getAuthorId().equals(userId)) {
                        return ResponseEntity.status(HttpStatus.FORBIDDEN).<Comment>build();
                    }
                    comment.setText(body.get("text"));
                    comment.setUpdatedAt(LocalDateTime.now());
                    return ResponseEntity.ok(commentRepository.save(comment));
                })
                .orElse(ResponseEntity.notFound().build());
    }
}