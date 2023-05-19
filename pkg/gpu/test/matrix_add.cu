#include <stdio.h>
#include <iostream>
#include "cuda_runtime.h"
#include "device_launch_parameters.h"
#define get_tid_x() (blockIdx.x * blockDim.x + threadIdx.x)
#define get_tid_y() (blockIdx.y * blockDim.y + threadIdx.y)

const int M = 8;
const int N = 8;

static void HandleError(cudaError_t err,const char *file, int line) {
    if (err != cudaSuccess) {
        printf("%s in %s at line %d\n", cudaGetErrorString(err), file, line);
        cout<<endl;
        exit(EXIT_FAILURE);
    }
}

#define HANDLE_ERROR(err) (HandleError(err, __FILE__, __LINE__))

//核函数
__global__ void matrix_add(int **A, int **B, int **C) {
    int i = get_tid_x();
    int j = get_tid_y();  
    C[i][j] = A[i][j] + B[i][j];
}

int main() {
    int nbytes=M*N*sizeof(int);
    
    int **A = (int **) malloc(sizeof(int *) * M);
    int **B = (int **) malloc(sizeof(int *) * M);
    int **C = (int **) malloc(sizeof(int *) * M);

    int *data_A = (int *) malloc(sizeof(int) * M * N);
    int *data_B = (int *) malloc(sizeof(int) * M * N);
    int *data_C = (int *) malloc(sizeof(int) * M * N);


    //这里使用了强制类型转换
    int *host_A = (int *) malloc(nbytes);
    int *host_B = (int *) malloc(nbytes);
    int *host_C = (int *) malloc(nbytes);


    for (int i = 0; i < M * N; i++) {
        data_A[i] = i;
        data_B[i] = i;
        host_A[i] = i;
        host_B[i] = i;
    }

    printf("Matrix A is:\n");
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            printf("%d ", data_A[i * N + j]);
        }
        printf("\n");
    }

    printf("Matrix B is:\n");
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            printf("%d ", data_B[i * N + j]);
        }
        printf("\n");
    }

    int *dev_data_A;
    int *dev_data_B;
    int *dev_data_C;

    // malloc matrix (size = M*N) in GPU device
    HANDLE_ERROR(cudaMalloc((void **) &dev_data_A, sizeof(int) * M * N));
    HANDLE_ERROR(cudaMalloc((void **) &dev_data_B, sizeof(int) * M * N));
    HANDLE_ERROR(cudaMalloc((void **) &dev_data_C, sizeof(int) * M * N));

    // copy data from host to GPU device
    HANDLE_ERROR(cudaMemcpy((void *) dev_data_A, (void *) data_A, sizeof(int) * M * N, cudaMemcpyHostToDevice));
    HANDLE_ERROR(cudaMemcpy((void *) dev_data_B, (void *) data_B, sizeof(int) * M * N, cudaMemcpyHostToDevice));
    // init C
    HANDLE_ERROR(cudaMemset((void *) dev_data_C, 0, sizeof(int) * M * N));

    for (int i = 0; i < M; i++) {
        A[i] = dev_data_A + i * N;
        B[i] = dev_data_B + i * N;
        C[i] = dev_data_C + i * N;
    }

    int **dev_A;
    int **dev_B;
    int **dev_C;

    HANDLE_ERROR(cudaMalloc((void **) &dev_A, sizeof(int *) * M));
    HANDLE_ERROR(cudaMalloc((void **) &dev_B, sizeof(int *) * M));
    HANDLE_ERROR(cudaMalloc((void **) &dev_C, sizeof(int *) * M));

    HANDLE_ERROR(cudaMemcpy((void *) dev_A, (void *) A, sizeof(int *) * M, cudaMemcpyHostToDevice));
    HANDLE_ERROR(cudaMemcpy((void *) dev_B, (void *) B, sizeof(int *) * M, cudaMemcpyHostToDevice));
    HANDLE_ERROR(cudaMemcpy((void *) dev_C, (void *) C, sizeof(int *) * M, cudaMemcpyHostToDevice));

    dim3 threadPerBlock(5, 5);
    dim3 numBlocks(M / threadPerBlock.x, N / threadPerBlock.y);

    matrix_add <<<numBlocks, threadPerBlock>>> (dev_A, dev_B, dev_C);

    // copy result to host
    HANDLE_ERROR(cudaMemcpy((void *) data_C, (void *) dev_data_C, sizeof(int) * M * N, cudaMemcpyDeviceToHost));

    // print result: 
    printf("The matrix add result is:\n");
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            printf("%d ", data_C[i * N + j]);
        }
        printf("\n");
    }
}