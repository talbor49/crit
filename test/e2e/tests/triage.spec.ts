import { test, expect, type Page } from '@playwright/test';
import { clearAllComments, loadPage, mdSection, switchToDocumentView, addComment, getMdPath } from './helpers';

// ============================================================
// Fix / Skip triage buttons + severity headlines + review brief
// ============================================================
test.describe('Comment triage', () => {
  test.beforeEach(async ({ page, request }) => {
    await clearAllComments(request);
  });

  test('Fix posts a "fix" reply without typing', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, '[Critical] first finding');
    await loadPage(page);
    await switchToDocumentView(page);

    const card = mdSection(page).locator('.comment-card').first();
    await expect(card).toBeVisible();
    await card.locator('.btn-fix').click();

    await expect(card.locator('.reply-body')).toContainText('fix');
  });

  test('Skip posts a "skip" reply', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, '[Suggestion] rename this');
    await loadPage(page);
    await switchToDocumentView(page);

    const card = mdSection(page).locator('.comment-card').first();
    await expect(card).toBeVisible();
    await card.locator('.btn-skip').click();

    await expect(card.locator('.reply-body')).toContainText('skip');
  });

  // The nav highlight self-clears after 1s, which is too tight to assert on
  // reliably; what matters here is that triaging one comment leaves the rest
  // of the queue interactive rather than stuck disabled.
  test('triaging one comment leaves the next one triageable', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, '[Critical] first finding');
    await addComment(request, mdPath, 3, '[Important] second finding');
    await loadPage(page);
    await switchToDocumentView(page);

    const cards = mdSection(page).locator('.comment-card');
    await expect(cards).toHaveCount(2);
    await cards.first().locator('.btn-fix').click();
    await expect(cards.first().locator('.reply-body')).toContainText('fix');

    await cards.nth(1).locator('.btn-skip').click();
    await expect(cards.nth(1).locator('.reply-body')).toContainText('skip');
  });
});

test.describe('Comment severity', () => {
  test.beforeEach(async ({ request }) => {
    await clearAllComments(request);
  });

  test('a [Critical] tag renders as a badge and leaves the body clean', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, '[Critical] Deadlocks on a slow IdP\n\nHolding the lock across the call.');
    await loadPage(page);
    await switchToDocumentView(page);

    const card = mdSection(page).locator('.comment-card').first();
    await expect(card).toHaveClass(/severity-critical/);
    await expect(card.locator('.comment-severity-badge')).toHaveText('Critical');
    await expect(card.locator('.comment-severity-title')).toHaveText('Deadlocks on a slow IdP');
    await expect(card.locator('.comment-body')).toContainText('Holding the lock across the call.');
    await expect(card.locator('.comment-body')).not.toContainText('[Critical]');
  });

  test('each severity gets its own class', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, '[Important] second');
    await addComment(request, mdPath, 3, '[Suggestion] third');
    await addComment(request, mdPath, 5, '[Missing test] fourth');
    await loadPage(page);
    await switchToDocumentView(page);

    const section = mdSection(page);
    await expect(section.locator('.comment-card.severity-important')).toHaveCount(1);
    await expect(section.locator('.comment-card.severity-suggestion')).toHaveCount(1);
    await expect(section.locator('.comment-card.severity-missing-test')).toHaveCount(1);
  });

  test('an untagged comment gets no severity treatment', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, 'plain note, no tag');
    await loadPage(page);
    await switchToDocumentView(page);

    const card = mdSection(page).locator('.comment-card').first();
    await expect(card).not.toHaveClass(/has-severity/);
    await expect(card.locator('.comment-severity-badge')).toHaveCount(0);
  });
});

// ============================================================
// One-click quick comments in the composer
// ============================================================
test.describe('Quick comment actions', () => {
  test.beforeEach(async ({ page, request }) => {
    await clearAllComments(request);
    await loadPage(page);
    await switchToDocumentView(page);
  });

  async function openCommentForm(page: Page) {
    const section = mdSection(page);
    const lineBlock = section.locator('.line-block').first();
    await lineBlock.scrollIntoViewIfNeeded();
    await lineBlock.hover();
    await section.locator('.line-comment-gutter').first().click();
    await expect(page.locator('.comment-form textarea')).toBeVisible();
  }

  test('the composer offers the three quick actions', async ({ page }) => {
    await openCommentForm(page);
    await expect(page.locator('.comment-form .quick-action')).toHaveText(['Fix it', 'Explain it', 'PR comment']);
  });

  test('one click posts the canned body with no typing', async ({ page }) => {
    await openCommentForm(page);
    await page.locator('.comment-form .quick-action', { hasText: 'Fix it' }).click();

    const card = mdSection(page).locator('.comment-card').first();
    await expect(card).toBeVisible();
    await expect(card.locator('.comment-body')).toHaveText('fix it');
    // The form submits and closes — the whole point is a single click.
    await expect(page.locator('.comment-form textarea')).toHaveCount(0);
  });

  test('each action posts its own body', async ({ page }) => {
    for (const [label, body] of [['Explain it', 'explain it'], ['PR comment', 'comment inline on the PR for me as a separate comment']]) {
      await openCommentForm(page);
      await page.locator('.comment-form .quick-action', { hasText: label }).click();
      await expect(mdSection(page).locator('.comment-card .comment-body').last()).toHaveText(body);
    }
  });

  // A misclick must not throw away what the reviewer already wrote.
  test('typed text is kept and the directive appended', async ({ page }) => {
    await openCommentForm(page);
    await page.locator('.comment-form textarea').fill('this breaks on empty tenants');
    await page.locator('.comment-form .quick-action', { hasText: 'Fix it' }).click();

    const body = mdSection(page).locator('.comment-card .comment-body').first();
    await expect(body).toContainText('this breaks on empty tenants');
    await expect(body).toContainText('fix it');
  });

  test('editing an existing comment offers no quick actions', async ({ page, request }) => {
    const mdPath = await getMdPath(request);
    await addComment(request, mdPath, 1, 'original text');
    await loadPage(page);
    await switchToDocumentView(page);

    const section = mdSection(page);
    await section.locator('.comment-actions button[title="Edit"]').first().click();
    await expect(section.locator('textarea')).toBeVisible();
    await expect(page.locator('.comment-form .quick-action')).toHaveCount(0);
  });
});
