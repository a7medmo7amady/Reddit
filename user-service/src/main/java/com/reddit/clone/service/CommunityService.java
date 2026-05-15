package com.reddit.clone.service;

import com.reddit.clone.dto.CommunityDTO;
import com.reddit.clone.dto.CreateCommunityRequest;
import com.reddit.clone.event.CommunityCreatedEvent;
import java.time.format.DateTimeFormatter;
import com.reddit.clone.exception.UserNotFoundException;
import com.reddit.clone.model.Community;
import com.reddit.clone.model.CommunityDocument;
import com.reddit.clone.model.User;
import com.reddit.clone.model.UserCommunity;
import com.reddit.clone.repository.CommunityDocumentRepository;
import com.reddit.clone.repository.CommunityRepository;
import com.reddit.clone.repository.UserCommunityRepository;
import com.reddit.clone.repository.UserRepository;
import org.springframework.cache.annotation.CacheEvict;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class CommunityService {

    private final CommunityRepository communityRepository;
    private final CommunityDocumentRepository communityDocumentRepository;
    private final UserRepository userRepository;
    private final UserCommunityRepository userCommunityRepository;
    private final KafkaEventPublisher kafkaEventPublisher;

    public CommunityService(CommunityRepository communityRepository,
                            CommunityDocumentRepository communityDocumentRepository,
                            UserRepository userRepository,
                            UserCommunityRepository userCommunityRepository,
                            KafkaEventPublisher kafkaEventPublisher) {
        this.communityRepository = communityRepository;
        this.communityDocumentRepository = communityDocumentRepository;
        this.userRepository = userRepository;
        this.userCommunityRepository = userCommunityRepository;
        this.kafkaEventPublisher = kafkaEventPublisher;
    }

    @CacheEvict(value = "user-communities", key = "#creatorId")
    @Transactional
    public CommunityDTO createCommunity(Long creatorId, CreateCommunityRequest req) {
        if (communityRepository.findByName(req.name()).isPresent()) {
            throw new IllegalArgumentException("Community r/" + req.name() + " already exists");
        }

        User creator = userRepository.findById(creatorId)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + creatorId));

        // Persist to PostgreSQL
        Community community = new Community(req.name(), req.description(), creator);
        community = communityRepository.save(community);

        // Auto-join creator
        userCommunityRepository.save(new UserCommunity(creator, community));

        // Persist to MongoDB
        CommunityDocument doc = new CommunityDocument(
                community.getId(), community.getName(), community.getDescription(),
                creator.getId(), creator.getUsername());
        communityDocumentRepository.save(doc);

        // Publish Kafka event for search indexing
        CommunityCreatedEvent event = new CommunityCreatedEvent(
                doc.getId(), community.getId(), community.getName(),
                community.getDescription(), 1, creator.getUsername(),
                community.getCreatedAt().format(DateTimeFormatter.ISO_LOCAL_DATE_TIME));
        kafkaEventPublisher.publish("community.created", String.valueOf(community.getId()), event);

        return new CommunityDTO(community.getId(), community.getName(), community.getDescription(), 1);
    }

    @Cacheable(value = "user-communities", key = "#userId")
    @Transactional(readOnly = true)
    public List<CommunityDTO> getFollowedCommunities(Long userId) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + userId));
        return userCommunityRepository.findByUser(user).stream()
                .map(uc -> {
                    Community c = uc.getCommunity();
                    return new CommunityDTO(c.getId(), c.getName(), c.getDescription(), c.getMemberCount());
                })
                .toList();
    }

    @CacheEvict(value = "user-communities", key = "#userId")
    @Transactional
    public void joinCommunity(Long userId, String communityName) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + userId));
        Community community = communityRepository.findByName(communityName)
                .orElseThrow(() -> new IllegalArgumentException("Community not found: " + communityName));

        if (!userCommunityRepository.existsByUserAndCommunity(user, community)) {
            userCommunityRepository.save(new UserCommunity(user, community));
            community.setMemberCount(community.getMemberCount() + 1);
            communityRepository.save(community);
            communityDocumentRepository.findByName(communityName).ifPresent(doc -> {
                doc.setMemberCount(doc.getMemberCount() + 1);
                communityDocumentRepository.save(doc);
            });
        }
    }

    @CacheEvict(value = "user-communities", key = "#userId")
    @Transactional
    public void leaveCommunity(Long userId, String communityName) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + userId));
        Community community = communityRepository.findByName(communityName)
                .orElseThrow(() -> new IllegalArgumentException("Community not found: " + communityName));

        userCommunityRepository.findByUserAndCommunity(user, community).ifPresent(uc -> {
            userCommunityRepository.delete(uc);
            community.setMemberCount(Math.max(0, community.getMemberCount() - 1));
            communityRepository.save(community);
            communityDocumentRepository.findByName(communityName).ifPresent(doc -> {
                doc.setMemberCount(Math.max(0, doc.getMemberCount() - 1));
                communityDocumentRepository.save(doc);
            });
        });
    }

    @Transactional(readOnly = true)
    public boolean isMember(Long userId, String communityName) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new UserNotFoundException("User not found: " + userId));
        Community community = communityRepository.findByName(communityName)
                .orElseThrow(() -> new IllegalArgumentException("Community not found: " + communityName));
        return userCommunityRepository.existsByUserAndCommunity(user, community);
    }

    @Transactional(readOnly = true)
    public CommunityDTO getCommunity(String name) {
        Community community = communityRepository.findByName(name)
                .orElseThrow(() -> new IllegalArgumentException("Community not found: " + name));
        return new CommunityDTO(community.getId(), community.getName(), community.getDescription(), community.getMemberCount());
    }
}
