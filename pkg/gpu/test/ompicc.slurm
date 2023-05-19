#!/bin/bash

#SBATCH --job-name=cuda
#SBATCH --partition=dgx2
#SBATCH --output=%j.out
#SBATCH --error=%j.err
#SBATCH -N 1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=6
#SBATCH --gres=gpu:1
ulimit -l unlimited
ulimit -s unlimited

module load cuda/9.2.88-gcc-4.8.5
./matrix_add
./matrix_multiply



#!/bin/bash

#SBATCH --job-name=cublas
#SBATCH --partition=dgx2
#SBATCH --output=%j.out
#SBATCH --error=%j.err
#SBATCH -N 1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=6
#SBATCH --gres=gpu:1

ulimit -s unlimited
ulimit -l unlimited

module load gcc/8.3.0-gcc-4.8.5 cuda/10.1.243-gcc-8.3.0

./cublashello
