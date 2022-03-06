package utils

// So, there are other more effective solutions.
// But I get this one, because I've made it.
func LevenshteinDistance(str1, str2 string) int {
	// [      s t r 1]
	// [    0 1 2 3 4]
	// [  0          ]
	// [s 1          ]
	// [t 2          ]
	// [r 3          ]
	// [2 4          ]

	str1len := len(str1)
	str2len := len(str2)

	matrix := make([][]int, str2len+1)

	// Fill downwards and set the cost of changing str2
	// to the empty string
	for i := 0; i <= str2len; i++ {
		matrix[i] = make([]int, str1len+1)
		matrix[i][0] = i
	}

	// Fill leftwards cost of changing str1
	// to the empty string
	for i := 0; i <= str1len; i++ {
		matrix[0][i] = i
	}

	// Go down str2 one row at a time and fill left
	// the cost of changing str1 to the substring of str2
	for i := 1; i <= str2len; i++ {
		for j := 1; j <= str1len; j++ {
			if str1[j-1] == str2[i-1] {
				matrix[i][j] = matrix[i-1][j-1]
				continue
			}

			min := matrix[i-1][j]
			if matrix[i][j-1] < min {
				min = matrix[i][j-1]
			}

			if matrix[i-1][j-1] < min {
				min = matrix[i-1][j-1]
			}

			matrix[i][j] = min + 1
		}
	}

	return matrix[str2len][str1len]
}
