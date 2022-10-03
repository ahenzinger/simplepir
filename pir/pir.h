#include <stdint.h>
#include <stddef.h>

typedef uint32_t Elem;

void transpose(Elem *out, const Elem *in, size_t rows, size_t cols);

void matMul(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t bCols);

void matMulTransposedPacked(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t bRows, size_t bCols);

void matMulVec(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols);

void matMulVecPacked(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols);
