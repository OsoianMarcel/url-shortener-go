package command

import (
	"errors"
	"fmt"

	"github.com/OsoianMarcel/url-shortener/internal/domain"
	"github.com/spf13/cobra"
)

type shortCommand struct {
	shortLinkUsecase domain.ShortLinkUsecase
}

func NewShortCommand(shortLinkUsecase domain.ShortLinkUsecase) *cobra.Command {
	h := &shortCommand{
		shortLinkUsecase: shortLinkUsecase,
	}

	cmd := &cobra.Command{
		Use:   "short",
		Short: "Manage short links",
	}

	cmd.AddCommand(
		h.newCreateCommand(),
		h.newExpandCommand(),
		h.newDeleteCommand(),
	)

	return cmd
}

func (h *shortCommand) newCreateCommand() *cobra.Command {
	var originalURL string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create short URL",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out, err := h.shortLinkUsecase.Create(cmd.Context(), domain.CreateAction{OriginalURL: originalURL})
			if err != nil {
				return mapShortLinkError(err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Key: %s\nShort URL: %s\n", out.Key, out.ShortURL)
			return err
		},
	}

	cmd.Flags().StringVar(&originalURL, "url", "", "Original URL to shorten")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func (h *shortCommand) newExpandCommand() *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "expand",
		Short: "Expand key to original URL",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ent, err := h.shortLinkUsecase.Expand(cmd.Context(), key)
			if err != nil {
				return mapShortLinkError(err)
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Original URL: %s\n", ent.OriginalURL)
			return err
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Short URL key")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func (h *shortCommand) newDeleteCommand() *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete short URL by key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := h.shortLinkUsecase.Delete(cmd.Context(), key); err != nil {
				return mapShortLinkError(err)
			}

			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Deleted key: %s\n", key)
			return err
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Short URL key")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func mapShortLinkError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidURL):
		return errors.New("invalid URL")
	case errors.Is(err, domain.ErrShortLinkNotFound):
		return errors.New("link not found")
	default:
		return fmt.Errorf("internal error: %w", err)
	}
}
