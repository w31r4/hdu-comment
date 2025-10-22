package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/common"
	"github.com/hdu-dp/backend/internal/dto"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/storage"
	"gorm.io/gorm"
)

// ReviewService contains business logic around review workflows.
type ReviewService struct {
	reviews *repository.ReviewRepository
	storage storage.FileStorage
	db      *gorm.DB
}

// NewReviewService constructs a review service instance.
func NewReviewService(reviews *repository.ReviewRepository, fileStorage storage.FileStorage, db *gorm.DB) *ReviewService {
	return &ReviewService{reviews: reviews, storage: fileStorage, db: db}
}

// ListFilters describes filters sortable/paginatable lists.
type ListFilters struct {
	Page     int
	PageSize int
	Query    string
	SortBy   string
	SortDir  string
	StoreID  *uuid.UUID
}

// Pagination metadata for list responses.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ReviewListResult wraps review list responses with pagination info.
type ReviewListResult struct {
	Data       []dto.ReviewResponse `json:"data"`
	Pagination Pagination           `json:"pagination"`
}

// Submit creates a new review in pending state for a specific store.
func (s *ReviewService) Submit(authorID uuid.UUID, storeID uuid.UUID, req dto.CreateReviewRequest) (*models.Review, error) {
	// 检查用户是否已经评价过该店铺
	_, err := s.reviews.FindByUserAndStore(authorID, storeID)
	if err == nil {
		return nil, errors.New("user has already reviewed this store")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	review := &models.Review{
		StoreID:  storeID,
		Title:    req.Title,
		Content:  req.Content,
		Rating:   req.Rating,
		Status:   models.ReviewStatusPending,
		AuthorID: authorID,
	}

	if err := s.reviews.Create(review); err != nil {
		return nil, err
	}
	// Eager load author and images for the response
	return s.reviews.FindByID(review.ID)
}

// CreateReviewForNewStore creates a new store and a review for it.
func (s *ReviewService) CreateReviewForNewStore(ctx context.Context, authorID uuid.UUID, req dto.CreateReviewForNewStoreRequest) (*models.Store, *models.Review, error) {
	// This is a simplified version. In a real app, you'd use a proper store service.
	store := &models.Store{
		Name:        req.StoreName,
		Address:     req.StoreAddress,
		Status:      models.StoreStatusPending,
		CreatedBy:   authorID,
		AutoCreated: true, // Mark as auto-created
	}
	if err := s.db.Create(store).Error; err != nil {
		return nil, nil, err
	}

	review, err := s.Submit(authorID, store.ID, req.CreateReviewRequest)
	if err != nil {
		// Rollback store creation if review submission fails
		s.db.Delete(store)
		return nil, nil, err
	}

	return store, review, nil
}

// ListPublic returns approved reviews.
func (s *ReviewService) ListPublic(filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.Statuses = []models.ReviewStatus{models.ReviewStatusApproved}
	opts.StoreID = filters.StoreID
	return s.listWithPagination(opts, filters)
}

// ListByStore returns approved reviews for a specific store.
func (s *ReviewService) ListByStore(storeID uuid.UUID, filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.Statuses = []models.ReviewStatus{models.ReviewStatusApproved}
	opts.StoreID = &storeID
	return s.listWithPagination(opts, filters)
}

// ListByAuthor returns reviews submitted by the specified user.
func (s *ReviewService) ListByAuthor(authorID uuid.UUID, filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.AuthorID = &authorID
	return s.listWithPagination(opts, filters)
}

// ListPending returns pending reviews for admin review.
func (s *ReviewService) ListPending(filters ListFilters) (ReviewListResult, error) {
	opts := buildListOptions(filters)
	opts.Statuses = []models.ReviewStatus{models.ReviewStatusPending}
	return s.listWithPagination(opts, filters)
}

func buildListOptions(filters ListFilters) repository.ListOptions {
	limit := filters.PageSize
	if limit <= 0 {
		limit = 10
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	return repository.ListOptions{
		Query:   filters.Query,
		SortBy:  filters.SortBy,
		SortDir: filters.SortDir,
		Limit:   limit,
		Offset:  offset,
	}
}

func (s *ReviewService) listWithPagination(opts repository.ListOptions, filters ListFilters) (ReviewListResult, error) {
	result, err := s.reviews.List(opts)
	if err != nil {
		return ReviewListResult{}, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}

	totalPages := int((result.Total + int64(limit) - 1) / int64(limit))

	return ReviewListResult{
		Data: dto.ToReviewListResponse(result.Reviews),
		Pagination: Pagination{
			Page:       page,
			PageSize:   limit,
			Total:      result.Total,
			TotalPages: totalPages,
		},
	}, nil
}

// Get returns a review by ID.
func (s *ReviewService) Get(id uuid.UUID) (*models.Review, error) {
	return s.reviews.FindByID(id)
}

// Update updates an existing review.
func (s *ReviewService) Update(ctx context.Context, userID, reviewID uuid.UUID, req dto.UpdateReviewRequest) (*models.Review, error) {
	review, err := s.reviews.FindByID(reviewID)
	if err != nil {
		return nil, errors.New("review not found")
	}

	if review.AuthorID != userID {
		return nil, errors.New("not authorized to update this review")
	}

	if req.Title != nil {
		review.Title = *req.Title
	}
	if req.Content != nil {
		review.Content = *req.Content
	}
	if req.Rating != nil {
		review.Rating = *req.Rating
	}

	// Updating a review should reset its status to pending for re-approval
	review.Status = models.ReviewStatusPending

	if err := s.reviews.Update(review); err != nil {
		return nil, err
	}

	// After updating, we might need to recalculate store's average rating
	// This can be done asynchronously or in a separate job
	// For now, we'll just return the updated review
	return s.reviews.FindByID(reviewID)
}

// Approve marks a review as approved.
func (s *ReviewService) Approve(review *models.Review) error {
	if review.Status != models.ReviewStatusPending {
		return common.ErrReviewAlreadyProcessed
	}
	review.Status = models.ReviewStatusApproved
	review.RejectionReason = ""
	return s.reviews.Update(review)
}

// Reject marks a review as rejected with reason.
func (s *ReviewService) Reject(review *models.Review, reason string) error {
	if review.Status != models.ReviewStatusPending {
		return common.ErrReviewAlreadyProcessed
	}
	review.Status = models.ReviewStatusRejected
	review.RejectionReason = strings.TrimSpace(reason)
	return s.reviews.Update(review)
}

// StoreImage saves the uploaded file via storage provider and records metadata.
func (s *ReviewService) StoreImage(ctx context.Context, reviewID uuid.UUID, file *storage.UploadFile) (*models.ReviewImage, error) {
	if file == nil {
		return nil, errors.New("file payload required")
	}

	defer file.Reader.Close()

	key := filepath.ToSlash(filepath.Join(reviewID.String(), fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(file.Filename))))

	info, err := s.storage.Save(ctx, key, file.Reader, file.Size, file.ContentType)
	if err != nil {
		return nil, err
	}

	image := &models.ReviewImage{
		ReviewID:   reviewID,
		StorageKey: info.Key,
		URL:        info.URL,
	}

	if err := s.reviews.AddImage(image); err != nil {
		return nil, err
	}

	return image, nil
}

// Delete removes a review by ID, ensuring ownership.
func (s *ReviewService) Delete(ctx context.Context, userID, reviewID uuid.UUID) error {
	review, err := s.reviews.FindByID(reviewID)
	if err != nil {
		return errors.New("review not found")
	}

	// Check ownership or admin role
	// This part is simplified, in a real app you'd get the user's role
	if review.AuthorID != userID {
		return errors.New("not authorized to delete this review")
	}

	// First, delete associated images from storage
	var keys []string
	if len(review.Images) > 0 {
		keys = make([]string, 0, len(review.Images))
		for _, image := range review.Images {
			if image.StorageKey != "" {
				keys = append(keys, image.StorageKey)
			}
		}
	}
	for _, key := range keys {
		if err := s.storage.Delete(ctx, key); err != nil {
			// Log error but continue, as the review itself is more important to delete
			fmt.Printf("failed to delete image %s from storage: %v\n", key, err)
		}
	}

	// Then, delete the review from the database
	return s.reviews.Delete(reviewID)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '-' || r == '.':
			return r
		default:
			return '_'
		}
	}, name)
}
