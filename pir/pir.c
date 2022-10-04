
#include "pir.h"
#include <stdio.h>

// Hard-coded, to allow for compiler optimizations:
#define COMPRESSION 3
#define BASIS       10
#define BASIS2      BASIS*2
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

void matMulTransposedPacked(Elem *out, const Elem *a, const Elem *b,
    size_t aRows, size_t aCols, size_t bRows, size_t bCols)
{
  Elem val, tmp, db;
  Elem tmp2, tmp3, tmp4, tmp5, tmp6, tmp7, tmp8;
  Elem db2, db3, db4, db5, db6, db7, db8;
  Elem val2, val3, val4, val5, vl6, val7, val8;
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
    for (size_t j = 0; j < bRows; j += 8) {
      ind1 = 0;
      for (size_t i = 0; i < aRows; i += 1) {
        tmp = 0;
	tmp2 = 0;
        tmp3 = 0;
	tmp4 = 0;
	tmp5 = 0;
	tmp6 = 0;
	tmp7 = 0;
	tmp8 = 0;
        ind2 = 0;
        for (size_t k = 0; k < aCols; k += 1) {
          db = a[ind1++];
          for (int m = 0; m < COMPRESSION; m++) {
            val = (db >> (m*BASIS)) & MASK;
            tmp += val*b[ind2+(j+0)*bCols];
            tmp2 += val*b[ind2+(j+1)*bCols];
            tmp3 += val*b[ind2+(j+2)*bCols];
            tmp4 += val*b[ind2+(j+3)*bCols];
            tmp5 += val*b[ind2+(j+4)*bCols];
            tmp6 += val*b[ind2+(j+5)*bCols];
            tmp7 += val*b[ind2+(j+6)*bCols];
            tmp8 += val*b[ind2+(j+7)*bCols];
            ind2++;
          }
        }
        out[bRows*i+j+0] = tmp;
        out[bRows*i+j+1] = tmp2;
        out[bRows*i+j+2] = tmp3;
        out[bRows*i+j+3] = tmp4;
        out[bRows*i+j+4] = tmp5;
        out[bRows*i+j+5] = tmp6;
        out[bRows*i+j+6] = tmp7;
        out[bRows*i+j+7] = tmp8;
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
  Elem db, db2, db3, db4, db5, db6, db7, db8;
  Elem val, val2, val3, val4, val5, val6, val7, val8;
  Elem tmp, tmp2, tmp3, tmp4, tmp5, tmp6, tmp7, tmp8;
  size_t index = 0;
  size_t index2;

  for (size_t i = 0; i < aRows; i += 8) {
    tmp  = 0;
    tmp2 = 0;
    tmp3 = 0;
    tmp4 = 0;
    tmp5 = 0;
    tmp6 = 0;
    tmp7 = 0;
    tmp8 = 0;

    index2 = 0;
    for (size_t j = 0; j < aCols; j++) {
      db  = a[index];
      db2 = a[index+1*aCols];
      db3 = a[index+2*aCols];
      db4 = a[index+3*aCols];
      db5 = a[index+4*aCols];
      db6 = a[index+5*aCols];
      db7 = a[index+6*aCols];
      db8 = a[index+7*aCols];

      val  = db & MASK;
      val2 = db2 & MASK;
      val3 = db3 & MASK;
      val4 = db4 & MASK;
      val5 = db5 & MASK;
      val6 = db6 & MASK;
      val7 = db7 & MASK;
      val8 = db8 & MASK;
      tmp  += val*b[index2];
      tmp2 += val2*b[index2];
      tmp3 += val3*b[index2];
      tmp4 += val4*b[index2];
      tmp5 += val5*b[index2];
      tmp6 += val6*b[index2];
      tmp7 += val7*b[index2];
      tmp8 += val8*b[index2];
      index2 += 1;

      val  = (db >> BASIS) & MASK;
      val2 = (db2 >> BASIS) & MASK;
      val3 = (db3 >> BASIS) & MASK;
      val4 = (db4 >> BASIS) & MASK;
      val5 = (db5 >> BASIS) & MASK;
      val6 = (db6 >> BASIS) & MASK;
      val7 = (db7 >> BASIS) & MASK;
      val8 = (db8 >> BASIS) & MASK;
      tmp  += val*b[index2];
      tmp2 += val2*b[index2];
      tmp3 += val3*b[index2];
      tmp4 += val4*b[index2];
      tmp5 += val5*b[index2];
      tmp6 += val6*b[index2];
      tmp7 += val7*b[index2];
      tmp8 += val8*b[index2];
      index2 += 1;

      val  = (db >> BASIS2) & MASK;
      val2 = (db2 >> BASIS2) & MASK;
      val3 = (db3 >> BASIS2) & MASK;
      val4 = (db4 >> BASIS2) & MASK;
      val5 = (db5 >> BASIS2) & MASK;
      val6 = (db6 >> BASIS2) & MASK;
      val7 = (db7 >> BASIS2) & MASK;
      val8 = (db8 >> BASIS2) & MASK;
      tmp  += val*b[index2];
      tmp2 += val2*b[index2];
      tmp3 += val3*b[index2];
      tmp4 += val4*b[index2];
      tmp5 += val5*b[index2];
      tmp6 += val6*b[index2];
      tmp7 += val7*b[index2];
      tmp8 += val8*b[index2];
      index2 += 1;
      index += 1;
    }
    out[i]   += tmp;
    out[i+1] += tmp2;
    out[i+2] += tmp3;
    out[i+3] += tmp4;
    out[i+4] += tmp5;
    out[i+5] += tmp6;
    out[i+6] += tmp7;
    out[i+7] += tmp8;
    index += aCols*7;
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
