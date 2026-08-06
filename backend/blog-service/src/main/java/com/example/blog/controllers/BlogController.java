package com.example.blog.controllers;

import com.example.blog.dto.BlogDto.CreateBlogRequest;
import com.example.blog.models.Blog;
import com.example.blog.repositories.BlogRepository;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.web.client.RestTemplate;
import org.springframework.web.client.RestClientException;
import java.io.File;
import java.util.Map;
import java.util.ArrayList;

import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/blogs")
@CrossOrigin(origins = "*")
public class BlogController {

    private final BlogRepository blogRepository;
    private final RestTemplate restTemplate;

    @Value("${app.base-url:http://localhost:8082}")
    private String baseUrl;

    @Value("${follower-service.url:http://localhost:8084}")
    private String followerServiceUrl;

    public BlogController(BlogRepository blogRepository, RestTemplate restTemplate) {
        this.blogRepository = blogRepository;
        this.restTemplate = restTemplate;
    }

    @PostMapping
    public ResponseEntity<Blog> create(@RequestBody CreateBlogRequest req,
                                       @RequestHeader("Authorization") String authHeader) {
        Long authorId = extractUserIdFromToken(authHeader);
        String username = extractUsernameFromToken(authHeader);

        Blog blog = new Blog();
        blog.setAuthorId(authorId);
        blog.setAuthorUsername(username);
        blog.setTitle(req.getTitle());
        blog.setDescription(req.getDescription());
        blog.setImages(req.getImages() != null ? req.getImages() : List.of());
        blog.setCreatedAt(LocalDateTime.now());

        Blog saved = blogRepository.save(blog);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }

    @GetMapping
    public ResponseEntity<List<Blog>> getAll() {
        return ResponseEntity.ok(blogRepository.findAll());
    }

    @GetMapping("/my")
    public ResponseEntity<?> getMyBlogs(@RequestHeader("Authorization") String authHeader) {
        try {
            Long userId = extractUserIdFromToken(authHeader);
            List<Blog> myBlogs = blogRepository.findByAuthorId(userId);
            return ResponseEntity.ok(myBlogs);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping("/followed")
    public ResponseEntity<?> getFollowedBlogsOfFollowedUsers(@RequestHeader("Authorization") String authHeader) {
        try {
            Long userId = extractUserIdFromToken(authHeader);

            // Call follower-service to get list of followed users
            String url = followerServiceUrl + "/following/" + userId;
            ResponseEntity<Map[]> response = restTemplate.getForEntity(url, Map[].class);

            if (response.getBody() == null || response.getBody().length == 0) {
                return ResponseEntity.ok(List.of());
            }

            // Extract user IDs from the response
            List<Long> followedUserIds = new ArrayList<>();
            for (Map<?, ?> user : response.getBody()) {
                Object id = user.get("id");
                if (id != null) {
                    followedUserIds.add(Long.parseLong(id.toString()));
                }
            }

            // Get blogs from followed users
            if (followedUserIds.isEmpty()) {
                return ResponseEntity.ok(List.of());
            }

            List<Blog> blogs = blogRepository.findByAuthorIdIn(followedUserIds);
            return ResponseEntity.ok(blogs);
        } catch (RestClientException e) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(Map.of("error", "Follower service unavailable"));
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(Map.of("error", e.getMessage()));
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

    @PostMapping("/upload")
public ResponseEntity<Map<String, String>> uploadImage(@RequestParam("file") MultipartFile file) {
    try {
        String uploadDir = "uploads/";
        File dir = new File(uploadDir);
        if (!dir.exists()) dir.mkdirs();

        String filename = System.currentTimeMillis() + "_" + file.getOriginalFilename();
    File dest = new File(uploadDir + filename).getAbsoluteFile();
    file.transferTo(dest);

        String url = baseUrl + "/uploads/" + filename;
        return ResponseEntity.ok(Map.of("url", url));
    } catch (Exception e) {
        return ResponseEntity.status(500).body(Map.of("error", "Upload failed"));
    }
}

    @GetMapping("/{id}")
    public ResponseEntity<Blog> getById(@PathVariable String id) {
        return blogRepository.findById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
}