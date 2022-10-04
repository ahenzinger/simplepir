
#include "pir.h"
#include <stdio.h>

// Hard-coded, to allow for compiler optimizations:
#define COMPRESSION 3
#define BASIS       10
#define BASIS2      BASIS*2
#define MASK        (1<<BASIS)-1

#define UNROLLING   8

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

void matMulTransposedPacked(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t bRows, size_t bCols)
{
  Elem val, tmp, db, val2, val3;
  size_t ind1, ind2;

  if (aRows > aCols) { // when the database rows are long
    ind1 = 0;
    for (size_t i = 0; i < aRows; i += 1) {
      for (size_t k = 0; k < aCols; k += 1) {
        db = a[ind1++];
    	val = db & MASK;
    	val2 = (db >> BASIS) & MASK;
    	val3 = (db >> BASIS2) & MASK;
        for (size_t j = 0; j < bRows; j += 1) {
	  out[bRows*i+j] += val*b[k*COMPRESSION+j*bCols];
	  out[bRows*i+j] += val2*b[k*COMPRESSION+j*bCols+1];
	  out[bRows*i+j] += val3*b[k*COMPRESSION+j*bCols+2];
	}
      }
    }
  } else { // when the database rows are short
    for (size_t j = 0; j < bRows; j += UNROLLING) {
      //ind1 = 0;
      for (size_t i = 0; i < aRows; i += 1) {
	for (int j1 = 0; j1 < UNROLLING; j1++) {
          tmp = 0;
          ind2 = 0;
          for (size_t k = 0; k < aCols; k += 1) {
            db = a[i*aCols+k];
            for (int m = 0; m < COMPRESSION; m++) {
              val = (db >> (m*BASIS)) & MASK;
              tmp += val*b[ind2+(j+j1)*bCols];
              ind2++;
            }
          }
          out[bRows*i+j+j1] = tmp;
	}
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

void matMulVecPacked(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols)
{
  Elem db[UNROLLING] __attribute__ ((aligned (UNROLLING*32)));
  Elem val[UNROLLING] __attribute__ ((aligned (UNROLLING*32)));
  Elem tmp[UNROLLING] __attribute__ ((aligned (UNROLLING*32)));
  size_t index = 0;
  size_t index2;

  for (size_t i = 0; i < aRows; i += UNROLLING) {
    for (int c = 0; c < UNROLLING; c++) {
      tmp[c] = 0;
    }

    index2 = 0;
    for (size_t j = 0; j < aCols; j++) {
      for (int c = 0; c < UNROLLING; c++) {
        db[c] = a[index+c*aCols];
	val[c] = db[c] & MASK;
	tmp[c] += val[c]*b[index2];
	val[c] = (db[c] >> BASIS) & MASK;
	tmp[c] += val[c]*b[index2+1];
        val[c] = (db[c] >> BASIS2) & MASK;
        tmp[c] += val[c]*b[index2+2];
      }
      index2 += 3;
      index += 1;
    }
    for (int c = 0; c < UNROLLING; c++) {
      out[i+c] = tmp[c];
    }
    index += aCols*(UNROLLING-1);
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
