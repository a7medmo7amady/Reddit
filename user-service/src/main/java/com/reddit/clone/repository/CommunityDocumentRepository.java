package com.reddit.clone.repository;

import com.reddit.clone.model.CommunityDocument;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface CommunityDocumentRepository extends MongoRepository<CommunityDocument, String> {
    Optional<CommunityDocument> findByName(String name);
    boolean existsByName(String name);
}
