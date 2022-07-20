
#include "pir.h"
#include <stdio.h>

// Hard-coded, to allow for compiler optimizations:
#define COMPRESSION 3
#define BASIS       10
#define MASK        (1<<BASIS)-1

void matMul(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t bCols)
{
  for (size_t i = 0; i < aRows; i++) {
    for (size_t k = 0; k < aCols; k++) {
      for (size_t j = 0; j < bCols; j++) {
        out[bCols*i + j] += a[aCols*i + k]*b[bCols*k + j];
      }
    }
  }
}

void matMulVec(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols)
{
  Elem tmp;
  for (size_t i = 0; i < aRows; i++) {
    tmp = 0;
    for (size_t j = 0; j < aCols; j++) {
      tmp += a[aCols*i + j]*b[j];
    }
    out[i] = tmp;
  }
}

void matMulVecSub(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t sub)
{
  Elem tmp;
  Elem val;

  for (size_t i = 0; i < aRows; i++) {
    tmp = 0;
    for (size_t j = 0; j < aCols; j++) {
      for (int k = 0; k < COMPRESSION; k++) {
        val = (a[aCols*i + j] >> (k*BASIS)) & MASK;
        val -= sub;
        tmp += val*b[j*COMPRESSION+k];
      }
    }
    out[i] = tmp;
  }
}

void transpose(Elem *out, const Elem *in, size_t rows, size_t cols)
{
  for (size_t i = 0; i < rows; i++) {
    for (size_t j = 0; j < cols; j++) {
      out[j*rows+i] = in[i*cols+j];
    }
  }
}
