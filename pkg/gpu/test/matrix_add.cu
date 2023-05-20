#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include "cuda_runtime.h"
#include "device_launch_parameters.h"
using namespace std;
const int M = 8;
const int N = 8;

__global__ void matrix_add(int **A, int **B, int **C) {
    int i = (blockIdx.x * blockDim.x + threadIdx.x);
    int j = (blockIdx.y * blockDim.y + threadIdx.y);
    C[i][j] = A[i][j] + B[i][j];
}

int main() {
    int nbytes=M*N*sizeof(int);
    //这两个是位于host机上的
    int **host_A = (int **) malloc(M * sizeof(int *));
    int **host_B = (int **) malloc(M * sizeof(int *));
    int **host_C = (int **) malloc(M * sizeof(int *));
    int *data_A = (int *) malloc(nbytes);
    int *data_B = (int *) malloc(nbytes);
    int *data_C = (int *) malloc(nbytes);
    for (int i = 0; i < M; i++) {
        host_A[i] = &data_A[i * N];
        host_B[i] = &data_B[i * N];
        host_C[i] = &data_C[i * N];
        for (int j = 0; j < N ; j++) {
            data_A[i*N+j] = i*N+j;
            data_B[i*N+j] = i*N+j;
            data_C[i*N+j] = 0;
        }
    }

    //这里说明了在host上的指针组成的数组是好的
    cout<<"host上面的矩阵A:"<<endl;
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            cout<<host_A[i][j]<<" ";
        }
        cout<<endl;
    }

    cout<<"host上面的矩阵B:"<<endl;
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            cout<<host_B[i][j]<<" ";
        }
        cout<<endl;
    }

    int **dev_A, **dev_B, **dev_C;
    int *dev_A1, *dev_B1, *dev_C1;
    cudaMalloc((void **)&dev_A1, nbytes);
    cudaMalloc((void **)&dev_B1, nbytes);
    cudaMalloc((void **)&dev_C1, nbytes);
    //数据拷贝
    cudaMemcpy((void *)dev_A1, (void *)data_A, nbytes, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_B1, (void *)data_B, nbytes, cudaMemcpyHostToDevice);

    cudaMemset((void *)dev_C1, 0, nbytes);
    for (int i = 0; i < M; i++) {
        host_A[i] = dev_A1 + i * N;
        host_B[i] = dev_B1 + i * N;
        host_C[i] = dev_C1 + i * N;
    }

    cudaMalloc((void **)&dev_A, sizeof(int *) * M);
    cudaMalloc((void **)&dev_B, sizeof(int *) * M);
    cudaMalloc((void **)&dev_C, sizeof(int *) * M);

    cudaMemcpy((void *)dev_A, (void *)host_A, sizeof(int *) * M, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_B, (void *)host_B, sizeof(int *) * M, cudaMemcpyHostToDevice);
    cudaMemcpy((void *)dev_C, (void *)host_C, sizeof(int *) * M, cudaMemcpyHostToDevice);


    dim3 grid(M / 2, N / 2);
    dim3 block(2, 2);
    matrix_add<<<grid, block>>>(dev_A, dev_B, dev_C);

    cudaMemcpy((void *) data_C,(void *) dev_C1, nbytes, cudaMemcpyDeviceToHost);

    cout<<"矩阵加法的结果:"<<endl;
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N ; j++) {
            cout<<data_C[i*N+j]<<" ";
        }
        cout<<endl;
    }
    free(data_A);
    free(data_B);
    free(data_C);
    free(host_A);
    free(host_B);
    free(host_C);
    cudaFree(dev_A);
    cudaFree(dev_B);
    cudaFree(dev_C);
    cudaFree(dev_A1);
    cudaFree(dev_B1);
    cudaFree(dev_C1);

    return 0;
}