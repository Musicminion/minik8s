#!/bin/bash

#SBATCH --job-name=cuda
#SBATCH --partition=dgx2
#SBATCH --output=%j.out
#SBATCH --error=%j.err
#SBATCH -N 1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=6
#SBATCH --gres=gpu:1
ulimit -s unlimited
ulimit -l unlimited

module load cuda/9.2.88-gcc-4.8.5
nvcc matrix_add.cu -o matrix_add
nvcc matrix_multiply.cu -o matrix_multiply
./matrix_add
./matrix_multiply