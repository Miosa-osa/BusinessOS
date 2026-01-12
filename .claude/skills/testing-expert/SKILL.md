---
name: testing-expert
description: Expert in writing comprehensive tests for Go backend and SvelteKit frontend. Use when writing tests, fixing test failures, or when test files are involved.
allowed-tools: Read, Edit, Write, Bash
---

# BusinessOS Testing Expert

## Go Backend Testing

```go
func TestServiceMethod(t *testing.T) {
    // Arrange
    ctx := context.Background()
    service := NewService(mockRepo)

    // Act
    result, err := service.Method(ctx, input)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Frontend Testing (Vitest)

```typescript
test('component renders', () => {
  render(Component, { props: { title: 'Test' } });
  expect(screen.getByText('Test')).toBeInTheDocument();
});
```

## Running Tests

```bash
# Backend
cd desktop/backend-go && go test ./...

# Frontend
cd frontend && npm test
```
