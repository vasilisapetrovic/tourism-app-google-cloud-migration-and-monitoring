package com.example.blog.dto;

import java.util.List;

public class BlogDto {

    public static class CreateBlogRequest {
        private String title;
        private String description;
        private List<String> images;

        public String getTitle() { return title; }
        public void setTitle(String title) { this.title = title; }
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
        public List<String> getImages() { return images; }
        public void setImages(List<String> images) { this.images = images; }
    }

}