package util

// [副露數][巡目][副露之後的手切數(手切數為0即鳴牌那一巡)]
var tenpaiRate = [][][]float64{
	{
		// TODO: 默聽
		// 可以大緻近似為巡目數
	},
	{
		// 1 副露
		{0},
		{1.05}, // 1
		{1.72, 3.37},
		{2.90, 5.60, 8.53},
		{5.26, 9.34, 13.45, 17.86},
		{8.56, 14.42, 19.88, 25.19, 31.25}, // 5
		{13.05, 21.04, 27.89, 33.85, 40.21, 47.34},
		{18.34, 28.05, 35.71, 42.58, 48.71, 55.85, 63.21},
		{24.38, 35.07, 43.76, 51.24, 57.73, 63.48, 68.56, 74.68},
		{30.11, 41.25, 50.79, 58.60, 64.60, 70.18, 74.26, 78.25, 80.23},
		{36.12, 47.47, 57.50, 64.90, 71.13, 76.06, 79.35, 82.00, 83.50, 83.50}, // 10
		{41.39, 53.62, 62.82, 69.87, 75.56, 79.85, 83.07, 85.42, 87.92, 89.02},
		{46.38, 58.24, 66.75, 74.18, 80.05, 84.12, 87.14, 89.50, 91.41, 92.73},
		{49.21, 61.84, 70.55, 78.84, 82.92, 86.07, 88.99, 91.38, 93.44, 94.63, 96.42},
		{52.07, 63.94, 73.27, 81.35, 85.42, 88.85, 91.24, 93.14, 94.28},
		{56.89, 68.11, 75.01, 81.98, 84.55, 87.98, 90.60, 92.04, 93.63}, // 15
		{62.86, 73.24, 79.46, 83.20, 86.56, 90.76, 92.81},
		{67.42, 75.09, 80.07, 80.07, 80.07, 86.71},
		{69.02, 84.51, 84.51}, // 18
	},
	{
		// 2 副露
		{0},
		{0},
		{9.92}, // 2
		{12.10, 17.72},
		{17.54, 25.55, 32.77},
		{23.24, 32.99, 41.82, 53.61}, // 5
		{30.32, 41.77, 51.53, 62.77, 68.97},
		{37.55, 49.50, 60.05, 69.07, 74.40, 76.53},
		{43.52, 56.91, 67.17, 75.04, 79.34, 80.43, 88.26},
		{50.28, 63.58, 73.48, 80.34, 83.85, 84.45, 89.63, 89.63},
		{55.06, 69.27, 77.62, 84.93, 88.26, 89.04, 90.73, 93.82, 96.91}, // 10
		{62.00, 74.63, 82.16, 88.18, 91.51, 92.42, 94.17, 97.08, 98.54, 98.54},
		{66.04, 78.10, 85.15, 89.57, 93.22, 94.76, 95.51, 98.20},
		{70.83, 81.37, 88.29, 91.52, 94.40, 96.40, 97.20, 98.60},
		{72.52, 82.80, 89.17, 92.46, 94.97, 96.52, 96.52},
		{75.00, 85.45, 90.74, 94.57, 95.85, 96.68, 96.68, 96.68}, // 15
		{76.98, 85.15, 89.27, 93.44, 94.38, 95.18, 95.18},
		{81.56, 88.65, 91.89, 96.76, 96.76},
		{84.01}, // 18
	},
	{
		// 3 副露
		{0},
		{0},
		{0},
		{43.81}, // 3
		{52.07, 63.48},
		{61.44, 70.28, 87.26}, // 5
		{68.20, 77.06, 90.17},
		{72.56, 82.85, 88.56, 88.56},
		{77.58, 87.59, 90.93},
		{81.16, 90.22, 92.74, 97.58},
		{84.53, 91.37, 94.75}, // 10
		{86.23, 92.28, 96.31, 98.15},
		{87.26, 93.05, 96.53, 98.26, 99.99},
		{87.24, 94.11, 96.83},
		{87.35, 94.58, 95.93},
		{87.66, 93.15, 94.29}, //15
		{89.03, 91.77, 93.42, 93.42},
		{89.54, 89.54},
		{90.19}, //18
	},
}

// 用四麻的數據近似得到三麻的數據
// 更加精確的數據見 https://shikkaku.com/data_sanma_20
func GetTenpaiRate3(rate4 float64) (rate3 float64) {
	return rate4 * (2 - rate4/100)
}
