package com.reddit.clone.Controllers;

import com.reddit.clone.dto.PublicProfileResponse;
import com.reddit.clone.model.Community;
import com.reddit.clone.service.SearchService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/search")
public class SearchController {

    private final SearchService searchService;

    public SearchController(SearchService searchService) {
        this.searchService = searchService;
    }

    @GetMapping
    public ResponseEntity<Map<String, Object>> search(@RequestParam String q) {
        List<PublicProfileResponse> users = searchService.searchUsers(q);
        List<Community> communities = searchService.searchCommunities(q);

        Map<String, Object> response = new HashMap<>();
        response.put("users", users);
        response.put("communities", communities);
        response.put("query", q);

        return ResponseEntity.ok(response);
    }

    @GetMapping("/users")
    public ResponseEntity<List<PublicProfileResponse>> searchUsers(@RequestParam String q) {
        return ResponseEntity.ok(searchService.searchUsers(q));
    }

    @GetMapping("/communities")
    public ResponseEntity<List<Community>> searchCommunities(@RequestParam String q) {
        return ResponseEntity.ok(searchService.searchCommunities(q));
    }
}
